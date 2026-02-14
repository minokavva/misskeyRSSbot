---
name: implement-feature
description: このプロジェクトに新機能を実装する際に使用する。クリーンアーキテクチャの層構造に沿ったファイル配置、命名規則、実装パターンに従う。
---

# 新機能の実装

このプロジェクトはクリーンアーキテクチャを採用している。新機能の実装時は以下の手順と配置規則に従う。

## 層構造と依存方向

依存は常に内側(domain)に向かう。逆方向の依存は禁止。

- domain層: 他の層に依存しない
- application層: domain層のみに依存する
- infrastructure層: domain層に依存する
- interfaces層: domain層に依存する
- main.go: すべての具象型を生成し依存関係を注入する

## ファイル配置

| 層 | パス | 役割 |
|---|---|---|
| エンティティ | internal/domain/entity/ | ビジネスロジックを持つ構造体 |
| リポジトリIF | internal/domain/repository/ | インターフェース定義のみ |
| サービス | internal/application/ | ユースケースの実装 |
| 外部接続 | internal/infrastructure/{サービス名}/ | 外部APIやDBとの接続実装 |
| 設定 | internal/interfaces/config/ | 環境変数による設定管理 |
| 組立 | main.go | 依存関係の注入とアプリケーション起動 |

## エンティティの追加

パッケージはentity。構造体とコンストラクタをセットで定義する。コンストラクタはNew{型名}の形式にする。

## リポジトリインターフェースの追加

パッケージはrepository。メソッドの第一引数はcontext.Context。戻り値にはerrorを含める。インターフェースは小さく保つ。

## infrastructure実装の追加

domain/repositoryのインターフェースを実装する。構造体はエクスポートしない。コンストラクタはNew{型名}の形式でインターフェース型を返す。

## applicationサービスの拡張

既存のRSSFeedServiceに機能を追加する場合、新しいリポジトリをコンストラクタの引数に追加する。Functional Optionsパターンで任意設定を追加できる。

## main.goでの組立

新しいリポジトリの具象型を生成し、サービスのコンストラクタに渡す。

## コーディング規則

- エクスポートする識別子はPascalCase、プライベートはcamelCase
- レシーバー名は1から2文字の短縮形
- エラーはfmt.Errorfと%wでラップしコンテキストを付与する
- panicは使用禁止。エラーはerrorとして返す
- グローバル変数は使用禁止。依存性注入を使用する
- init関数は使用禁止。明示的な初期化を優先する
- コード内にコメントは記載しない。関数名と変数名で意図を表現する
