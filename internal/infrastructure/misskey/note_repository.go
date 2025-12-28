package misskey

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"misskeyRSSbot/internal/domain/entity"
	"misskeyRSSbot/internal/domain/repository"
)

type rateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

func newRateLimiter(maxTokens int, refillRate time.Duration) *rateLimiter {
	return &rateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (rl *rateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed / rl.refillRate)
	if tokensToAdd > 0 {
		rl.tokens = min(rl.tokens+tokensToAdd, rl.maxTokens)
		rl.lastRefill = now
	}

	if rl.tokens <= 0 {
		waitTime := rl.refillRate - (now.Sub(rl.lastRefill) % rl.refillRate)
		rl.mu.Unlock()

		timer := time.NewTimer(waitTime)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			rl.mu.Lock()
			rl.tokens = 1
			rl.lastRefill = time.Now()
			rl.tokens--
			rl.mu.Unlock()
			return nil
		}
	}

	rl.tokens--
	rl.mu.Unlock()
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type noteRepository struct {
	host        string
	authToken   string
	client      *http.Client
	rateLimiter *rateLimiter
}

type Config struct {
	Host           string
	AuthToken      string
	MaxRequests    int
	RefillInterval time.Duration
}

func NewNoteRepository(cfg Config) repository.NoteRepository {
	maxRequests := cfg.MaxRequests
	if maxRequests == 0 {
		maxRequests = 3
	}
	refillInterval := cfg.RefillInterval
	if refillInterval == 0 {
		refillInterval = 10 * time.Second
	}

	return &noteRepository{
		host:        cfg.Host,
		authToken:   cfg.AuthToken,
		client:      &http.Client{Timeout: 30 * time.Second},
		rateLimiter: newRateLimiter(maxRequests, refillInterval),
	}
}

func (r *noteRepository) Post(ctx context.Context, note *entity.Note) error {
	if err := r.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}

	notePayload := map[string]interface{}{
		"i":          r.authToken,
		"text":       note.Text,
		"visibility": string(note.Visibility),
	}

	payload, err := json.Marshal(notePayload)
	if err != nil {
		return fmt.Errorf("failed to serialize note: %w", err)
	}

	url := fmt.Sprintf("https://%s/api/notes/create", r.host)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Misskey API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Misskey API returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}
