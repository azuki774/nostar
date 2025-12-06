# ライブ配信機能のアーキテクチャ

## 主要コンポーネントの依存関係

```
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                            │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │  Connection     │  │ ConnectionPool  │  │ Subscription│  │
│  │  (interface)    │  │                 │  │             │  │
│  │                 │  │ - mu: RWMutex   │  │ - ID        │  │
│  │ + ID()          │  │ - conns: map    │  │ - Filters   │  │
│  │ + WriteJSON()   │  │                 │  │             │  │
│  │ + Close()       │  └─────────┬───────┘  └─────────────┘  │
│  └─────────────────┘           │                            │
│                                │                            │
│               ┌────────────────┼────────────────┐           │
│               │                │                │           │
└───────────────┼────────────────┼────────────────┘           │
                │                │                            │
                ▼                ▼                            ▼
┌───────────────┼────────────────┼─────────────────────────────┐
│   Transport   │   Usecase      │   Domain                     │
├───────────────┼────────────────┼─────────────────────────────┤
│ WebSocket-    │ RelayService   │ SubscriptionRegistry        │
│ Connection    │                │                             │
│ (implements   │ - connectionPool *ConnectionPool            │
│  Connection)  │ - registry      *SubscriptionRegistry       │
│               │                │ - subscriptions: map[ConnID][]Subscription │
└───────────────┼────────────────┼─────────────────────────────┘
```

## 依存関係の詳細

### 1. Connection Interface
- **定義場所**: `domain/connection.go`
- **実装**: `transport/websocket/connection.go` (WebSocketConnection)
- **役割**: 接続の抽象化（WebSocket以外の実装も可能）
- **依存先**: なし（interface）

### 2. ConnectionPool
- **定義場所**: `domain/connection.go`
- **役割**: 複数接続の一元管理
- **依存関係**:
  - `Connection` interface に依存（mapの値として使用）
  - スレッドセーフ（RWMutex使用）
- **メソッド**:
  - `Add(conn Connection)` - 接続追加
  - `Remove(id ConnectionID)` - 接続削除
  - `Get(id ConnectionID)` - 接続取得
  - `BroadcastTo(ids []ConnectionID, message interface{})` - 一斉配信

### 3. Subscription
- **定義場所**: `domain/subscription.go`
- **役割**: クライアントの購読条件を表現
- **依存関係**: なし（独立したドメインエンティティ）
- **構造**:
  - `ID string` - サブスクリプションID（クライアント指定）
  - `Filters []Filter` - フィルタ条件（OR条件）

## ライブ配信時のデータフロー

```
1. クライアント接続
   WebSocketConnection → ConnectionPool.Add()

2. サブスクリプション登録
   Subscription → SubscriptionRegistry.Register()

3. 新イベント受信
   Event → RelayService.HandleEvent()

4. フィルタマッチング
   SubscriptionRegistry.FindMatchingConnections() → []ConnectionID

5. イベント配信
   ConnectionPool.BroadcastTo(connectionIDs, eventMessage)
```

## 依存の方向性

- **Domain → Infrastructure**: 依存なし（interface使用）
- **Transport → Domain**: WebSocketConnection が Connection を実装
- **Usecase → Domain**: RelayService が ConnectionPool と Subscription を使用
- **循環依存なし**: クリーンアーキテクチャ遵守

## 利点

1. **テスト容易性**: Connection interface によりモック可能
2. **拡張性**: 新しい接続タイプ（HTTP/2など）に対応可能
3. **スレッドセーフティ**: ConnectionPool が並行アクセスを管理
4. **責務分離**: 各コンポーネントが単一責任原則に従う

---

*このアーキテクチャにより、ライブ配信機能を堅牢かつ拡張性高く実装可能*
