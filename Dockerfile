FROM golang:1.24-alpine AS builder

RUN apk add --no-cache tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o misskeyRSSbot .

FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=builder /app/misskeyRSSbot .
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

VOLUME /app/data

ENTRYPOINT ["./misskeyRSSbot"]
