# NIP-01 実装状況とTODO

## NIP-01 実装状況とTODO

### ✅ 実装済みの機能
1. **イベント構造と検証**: `domain.Event` に必要なフィールドを実装、go-nostrライブラリを使った署名検証
2. **WebSocketメッセージ処理**: `WireMessage` で EVENT/REQ/CLOSE を適切にパース
3. **フィルタ処理**: `domain.Filter` と `NewFilterFromRaw` でフィルタをパース、マッチングロジック実装
4. **サブスクリプション**: `domain.Subscription` で複数フィルタのOR条件を表現
5. **イベント保存**: GORMを使った `EventStore.Save` の実装
6. **基本的なクエリ**: ID, Authors, Kinds, Since, Until, Limit での検索
7. **WebSocket通信**: EVENT受信時のOK応答、REQ時のEVENT/EOSE送信
8. **ライブ配信機能**: 新規イベントのリアルタイム配信（`BroadcastToSubscribers`）
   - `SubscriptionRegistry` による接続ごとのサブスクリプション管理
   - `ConnectionPool` によるWebSocket接続の一元管理（ポインタ型によるスレッドセーフティ確保）
   - REQメッセージでのサブスクリプション登録、CLOSEメッセージでの解除
9. **テスト**: ドメイン層、インフラ層、ユースケース層の包括的なテスト

### ❌ 未実装の機能

#### 1. **タグ検索の実装** (優先度: 高)
**場所**: `/workspaces/nostar/internal/infrastructure/db/db.go:120`
**問題**: EventStore.Query でタグ検索が実装されていない
**影響**: `#e`, `#p`, `#t` などのタグフィルタが機能しない
**実装内容**:
- PostgreSQLのJSONBフィールドを使ったタグ検索
- クエリ例: `WHERE tags->>'e' IN (?)` またはJSON配列内の検索
- パフォーマンス考慮（インデックス作成）

#### 2. **エラーハンドリングの改善** (優先度: 中)
**場所**: `/workspaces/nostar/internal/transport/websocket/server.go:109-122`
**問題**: ドメインエラー（署名NG）と内部エラー（DB障害）の区別が不十分
**実装内容**:
- 専用エラー型の定義（`domain.ErrInvalidSignature` など）
- エラー種別による適切なNOTICE/OKメッセージの送信

#### 3. **WebSocketセキュリティ** (優先度: 低)
**場所**: `/workspaces/nostar/internal/transport/websocket/server.go:19-21`
**問題**: CheckOriginが常にtrueを返す
**実装内容**:
- 適切なオリジンチェックの設定
- 必要に応じてCORS設定

#### 4. **検索機能 (NIP-50)** (優先度: 低)
**場所**: `/workspaces/nostar/internal/infrastructure/db/db.go:120`
**問題**: EventStore.Query で `search` フィールドが実装されていない
**影響**: イベント本文の検索が機能しない（NIP-50の拡張機能）
**実装内容**:
- PostgreSQLのLIKE/ILIKE検索によるcontentフィールド検索
- パフォーマンス考慮（全文検索インデックスの検討）
- NIP-50準拠のオプション機能として実装

### 📋 実装順序の提案

1. **タグ検索の実装** - 基本機能の完成
2. **エラーハンドリングの改善** - 堅牢性の向上
3. **WebSocketセキュリティ** - 運用時の安全性確保

これにより、NIP-01の基本仕様に完全に準拠したNostrリレーが完成します。

（注: NIP-50の検索機能はオプションの実装として検討可能です）
