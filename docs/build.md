# ビルド手順

このドキュメントでは、nostarプロジェクトのバイナリビルドとDockerコンテナビルドの仕組みについて説明します。

## プロジェクト概要

nostarはGo言語で実装されたNostrプロトコルのリレーサーバーです。以下の主要コンポーネントで構成されています：

- **メインアプリケーション**: Nostrイベントの受信・保存・配信を行うWebSocketサーバー
- **データベース**: PostgreSQLを使用したイベント永続化
- **マイグレーション**: dbmateを使用したデータベーススキーマ管理
- **設定管理**: TOML形式の設定ファイル

## バイナリビルド

### ローカルビルド（Makefile使用）

プロジェクトルートで以下のコマンドを実行します：

```bash
# 依存関係のインストールとセットアップ
make setup

# コードチェック（フォーマット、テスト、静的解析）
make check

# バイナリビルド
make bin

# クリーンアップ
make clean
```

#### Makefileの詳細

```makefile
BINARY_NAME ?= nostar          # 出力バイナリ名
BUILD_DIR ?= bin              # 出力ディレクトリ
PKG ?= ./                     # ビルド対象パッケージ
GOOS ?= linux                 # ターゲットOS
GOARCH ?= amd64               # ターゲットアーキテクチャ
LDFLAGS ?= -s -w -extldflags '-static'  # リンカフラグ

# バイナリビルド
bin:
    mkdir -p $(BUILD_DIR)
    CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
        -trimpath -ldflags "$(LDFLAGS)" \
        -o $(BUILD_DIR)/$(BINARY_NAME) $(PKG)
```

#### ビルド特性

- **静的リンク**: CGO_ENABLED=0により完全な静的バイナリを生成
- **最適化**: -s -wフラグでデバッグ情報とシンボルテーブルを除去
- **クロスコンパイル**: GOOS/GOARCHでターゲットプラットフォームを指定可能
- **デフォルト**: Linux AMD64向けのバイナリを生成

### 直接のgo buildコマンド

Makefileを使用せずに直接ビルドする場合：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags "-s -w -extldflags '-static'" \
    -o bin/nostar .
```

## Dockerコンテナビルド

### メインアプリケーションコンテナ

`build/Dockerfile`を使用したマルチステージビルドで、以下の特徴を持つコンテナイメージを生成します：

#### ビルド特性

- **マルチステージビルド**: Goビルダーステージと最小限のランタイムステージを使用
- **セキュリティ**: 非特権ユーザーでの実行、最小限のベースイメージ
- **バージョン埋め込み**: Git情報に基づく動的バージョン情報埋め込み
- **静的リンク**: 完全な静的リンクによる移植性の高いバイナリ生成
- **ポート**: 9999番ポートを公開

#### ローカルDockerビルド

```bash
# 開発用イメージビルド
make build

# 直接docker build
docker build -t azuki774/nostar:dev -f build/Dockerfile .
```

### マイグレーションコンテナ

`scripts/migrations/Dockerfile`を使用し、データベースマイグレーション専用のコンテナイメージを生成します：

#### 特性

- **マイグレーションツール**: dbmateを使用した軽量なデータベーススキーマ管理
- **SQLファイル自動実行**: 指定ディレクトリのSQLファイルを順次実行
- **環境変数設定**: DATABASE_URLで接続先データベースを指定

## CI/CDパイプライン

### GitHub Actionsワークフロー

#### 1. Goワークフロー（.github/workflows/go.yml）

全ブランチプッシュ時に実行：

```yaml
- Go 1.25.4 セットアップ
- 依存関係ダウンロード
- make setup（ツールインストール）
- make check（フォーマット・テスト・静的解析）
- make bin（バイナリビルド）
```

#### 2. Dockerワークフロー（.github/workflows/docker.yml）

masterブランチとタグプッシュ時に実行：

- **appジョブ**: メインアプリケーションコンテナ
  - build/Dockerfileを使用
  - GitHub Container Registry (ghcr.io) にプッシュ
  - セマンティックバージョニングに基づくタグ付け

- **migrationジョブ**: マイグレーションコンテナ
  - scripts/migrations/Dockerfileを使用
  - 同様にghcr.ioにプッシュ

#### タグ付け戦略

```yaml
tags: |
  type=ref,event=branch      # ブランチ名
  type=ref,event=pr           # PR番号
  type=semver,pattern={{version}}        # v1.2.3
  type=semver,pattern={{major}}.{{minor}} # v1.2
  type=semver,pattern={{major}}           # v1
  type=raw,value=latest,enable={{is_default_branch}} # latest
```

#### 3. マイグレーションテスト（.github/workflows/migration-test.yml）

E2Eテスト実行：

- PostgreSQL 16コンテナ起動
- データベース初期化（001_init.sql）
- バイナリビルドとサーバー起動
- algia（Nostr CLIクライアント）を使用したテスト実行

### キャッシュ戦略

GitHub Actions Cacheを使用：

- Goモジュールキャッシュ: `/go/pkg/mod`
- Goビルドキャッシュ: `/root/.cache/go-build`
- Dockerレイヤーキャッシュ: GHA（GitHub Actions Cache）

## マイグレーション

### データベーススキーマ

マイグレーションSQLファイルにより、Nostrイベントを保存するためのテーブル構造とパフォーマンス用のインデックスが定義されます。主なテーブルにはイベント情報（ID、公開鍵、作成時刻、種別、タグ、内容、署名）と、リレー側のメタデータ（受信時刻、削除フラグ、隠しフラグなど）が含まれます。

### マイグレーション実行

#### ローカル実行

```bash
# dbmateインストール
go install github.com/amacneil/dbmate@latest

# マイグレーション実行
DATABASE_URL="postgres://user:pass@localhost:5432/nostar?sslmode=disable" \
  dbmate up
```

#### Dockerコンテナ使用

```bash
docker run --rm \
  -e DATABASE_URL="postgres://user:pass@localhost:5432/nostar?sslmode=disable" \
  -v $(pwd)/scripts/migrations/sql:/db/migrations \
  ghcr.io/amacneil/dbmate:2.28 \
  dbmate up
```

## 依存関係とバージョン

### Goバージョン
- **1.25.4**: プロジェクトで使用するGoバージョン


### 開発ツール

```bash
# インストールされるツール
go install honnef.co/go/tools/cmd/staticcheck@latest  # 静的解析
go install github.com/spf13/cobra-cli@latest          # Cobra CLIジェネレータ
go install github.com/amacneil/dbmate@latest          # DBマイグレーションツール
```

## 実行方法

### バイナリ実行

```bash
# 環境変数設定
export DATABASE_URL="postgres://user:pass@localhost:5432/nostar?sslmode=disable"

# サーバー起動
./bin/nostar serve -c ./bin/config.toml -p 9999
```

### Dockerコンテナ実行

```bash
# メインアプリケーション
docker run -d \
  --name nostar \
  -p 9999:9999 \
  -e DATABASE_URL="postgres://user:pass@host:5432/nostar?sslmode=disable" \
  -v $(pwd)/bin/config.toml:/config.toml \
  azuki774/nostar:latest

# マイグレーション
docker run --rm \
  -e DATABASE_URL="postgres://user:pass@host:5432/nostar?sslmode=disable" \
  azuki774/nostar-migration:latest
```

## 設定ファイル

### config.toml

```toml
[relay_info]
name = "nostar"
description = "A minimal Nostr relay implementation in Go"
pubkey = ""
contact = "admin@example.com"
supported_nips = [1, 11]
relay_countries = ["JP"]
language_tags = ["ja"]
posting_policy = "https://example.com/posting-policy"
```

この設定はNIP-11（Relay Information Document）に対応しています。

## まとめ

nostarプロジェクトのビルドシステムは以下の特徴を持ちます：

1. **シンプルなMakefile**: ローカル開発向けの基本的なビルドタスク
2. **マルチステージDockerビルド**: セキュリティと効率性を両立
3. **包括的なCI/CD**: GitHub Actionsによる自動化されたテストとデプロイ
4. **マイグレーションツール**: dbmateによるデータベーススキーマ管理
5. **クロスプラットフォーム対応**: 静的リンクによる移植性の高いバイナリ生成

これらの仕組みにより、開発から本番デプロイまでの一貫したワークフローが実現されています。
