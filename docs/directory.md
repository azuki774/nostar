# Nostr リレーサーバのディレクトリ構成（ヘキサゴナルアーキテクチャ）

```text
nostar/
├── bin/                         # ビルド成果物
├── cmd/                         # Cobra ベースの CLI コマンド群
│   ├── root.go                  # `nostar` コマンドのルート定義（Execute を提供）
│   └── serve.go                 # `nostar serve` サブコマンド（リレーサーバ起動を実装していく）
│
├── internal/
│   ├── relay/                   # Nostr リレーに関するドメイン＋ユースケース
│   │   ├── domain/              # イベント・サブスク等のドメインモデル（ビジネスルール）
│   │   │   ├── event.go         # Nostr イベント struct, 署名検証, バリデーション
│   │   │   ├── event_test.go    # イベント関連テスト
│   │   │   ├── filter.go        # フィルタ条件, REQ/CLOSE のモデルと判定ロジック
│   │   │   ├── filter_test.go   # フィルタ関連テスト
│   │   │   └── subscription.go  # サブスクリプションモデル
│   │   ├── usecase/             # ユースケース（サービス層 = インバウンドポート的な役割）
│   │   │   ├── messages.go      # メッセージ構造体定義
│   │   │   ├── relay_service.go # EVENT/REQ/CLOSE を受けて、ドメイン＋ポートを組み合わせて処理
│   │   │   └── relay_service_test.go # リレースサービステスト
│   │   └── port.go              # DB や PubSub への依存を抽象化した interface 群（アウトバウンドポート）
│   │
│   ├── infrastructure/          # 外部インフラ（DB, キャッシュ, 外部サービス等）の実装（アウトバウンドアダプター）
│   │   └── db/
│   │       ├── db.go            # PostgreSQL を使用したイベントストア実装
│   │       └── db_test.go       # データベーステスト（未実装）
│   │
│   ├── logger/                  # ロギング機能
│   │   └── logger.go
│   │
│   └── transport/               # 入出力のプロトコル層（インバウンドアダプター）
│       └── websocket/
│           ├── server.go        # Nostr WebSocket プロトコル実装（メッセージをパースして relay_service を呼ぶ）
│           └── wire.go          # WebSocket メッセージのワイヤーフォーマット
│
├── scripts/                     # ユーティリティスクリプト
├── docs/                        # ドキュメント
│   ├── directory.md             # このファイル
│   └── TODO.md                  # TODO リスト
│
├── LICENSE
├── Makefile                     # ビルドスクリプト
├── README.md
├── main.go                      # アプリケーションエントリ（cmd.Execute を呼び出すだけ）
├── go.mod
└── go.sum
```

上記の構成の説明:
- `relay/domain`: Nostr プロトコル上の「データ構造」と「ルール」を表す純粋なロジックを書く場所。イベント、フィルタ、サブスクリプションのモデルとビジネスロジックを含む。
- `relay/usecase`: WebSocket から来た EVENT/REQ/CLOSE を「どう処理するか」を組み立てるサービス層。ここから `port.go` の interface を呼び出す。
- `relay/port.go`: 「イベントを保存する」「イベントを検索する」など、インフラに依存する操作を interface で宣言する。
- `infrastructure/db`: PostgreSQL を使用して `relay/port.go` の EventStore interface を実装する。
- `transport/websocket`: WebSocket からの入出力を扱い、受け取ったリクエストを `usecase` に橋渡しする。
- `logger`: アプリケーション全体で使用するロギング機能を提供する。
