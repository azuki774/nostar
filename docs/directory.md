# Nostr リレーサーバのディレクトリ構成案（ヘキサゴナルアーキテクチャ）

```text
nostar/
├── cmd/                         # Cobra ベースの CLI コマンド群
│   ├── root.go                  # `nostar` コマンドのルート定義（Execute を提供）
│   └── serve.go                 # `nostar serve` サブコマンド（リレーサーバ起動を実装していく）
│
├── internal/
│   ├── relay/                   # Nostr リレーに関するドメイン＋ユースケース
│   │   ├── domain/              # イベント・サブスク等のドメインモデル（ビジネスルール）
│   │   │   ├── event.go         # Nostr イベント struct, 署名検証, バリデーション
│   │   │   ├── subscription.go  # フィルタ条件, REQ/CLOSE のモデルと判定ロジック
│   │   │   └── relay.go         # リレーとしてのポリシー（レート制限, ブラックリスト等）
│   │   ├── usecase/             # ユースケース（サービス層 = インバウンドポート的な役割）
│   │   │   ├── relay_service.go # EVENT/REQ/CLOSE を受けて、ドメイン＋ポートを組み合わせて処理
│   │   │   └── auth_service.go  # 公開鍵ごとの制限, 認可ポリシーなど
│   │   └── port.go              # DB や PubSub への依存を抽象化した interface 群（アウトバウンドポート）
│   │
│   ├── infrastructure/          # 外部インフラ（DB, キャッシュ, 外部サービス等）の実装（アウトバウンドアダプター）
│   │   ├── postgres/
│   │   │   ├── event_store.go   # relay.port.EventStore の Postgres 実装
│   │   │   └── migrations/
│   │   ├── redis/
│   │   │   └── pubsub.go        # relay.port.PubSub の Redis 実装
│   │   └── metrics/
│   │       └── prometheus.go    # メトリクスの実装（リクエスト数, レイテンシなど）
│   │
│   └── transport/               # 入出力のプロトコル層（インバウンドアダプター）
│       ├── http/
│       │   └── handler.go       # HTTP / REST エンドポイント（/health, /metrics, 管理APIなど）
│       └── websocket/
│           └── handler.go       # Nostr WebSocket プロトコル実装（メッセージをパースして relay_service を呼ぶ）
│
├── pkg/                         # 外部からも再利用可能な汎用ライブラリ
│
├── config/                      # アプリケーション設定ファイル（YAML/JSON など）
│   └── default.yml
│
├── deploy/                      # デプロイ用マニフェスト（Docker, k8s など）
│   ├── docker-compose.yml
│   └── k8s/
│
├── .github/
│   └── workflows/
│       └── migration-test.yml   # 既存のマイグレーションテスト
│
├── .devcontainer/
│   └── devcontainer.json        # Dev Container 設定
│
├── docs/
│   └── directory.md             # このファイル
│
├── main.go                      # アプリケーションエントリ（cmd.Execute を呼び出すだけ）
├── go.mod
└── go.sum
```

上記のイメージ:
- `relay/domain`: Nostr プロトコル上の「データ構造」と「ルール」を表す純粋なロジックを書く場所。
- `relay/usecase`: WebSocket から来た EVENT/REQ/CLOSE を「どう処理するか」を組み立てるサービス層。ここから `port.go` の interface を呼び出す。
- `relay/port.go`: 「イベントを保存する」「購読中のクライアントに通知する」など、インフラに依存する操作を interface で宣言する。
- `infrastructure/*`: Postgres/Redis/Prometheus など具体的なミドルウェアを使って `port.go` の interface を実装する。
- `transport/*`: HTTP や WebSocket からの入出力を扱い、受け取ったリクエストを `usecase` に橋渡しする。
