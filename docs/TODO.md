# TODO
ここでは、実装が中途半端になっているもの（実装されているように見えるが、正しく動かないものなど）を列挙する。

## データベース機能
- [ ] **タグ検索の実装**: `internal/infrastrcture/db/db.go:121` で TODO コメントあり。データベースでのタグベースのイベント検索が未実装。

## WebSocket セキュリティと機能
- [ ] **CheckOrigin のセキュリティ改善**: `internal/transport/websocket/server.go:20` で常に `true` を返す実装。CORS セキュリティが不十分。
- [ ] **複数フィルター対応**: `internal/transport/websocket/server.go:129,139` で TODO コメントあり。REQ メッセージで複数のフィルターを処理できない。
- [ ] **WebSocket アップグレード/ループ実装**: `internal/transport/websocket/server.go:39` でスケルトン状態。実際の WebSocket 処理が不完全。

## テスト
- [ ] **データベーステストの実装**: `internal/infrastrcture/db/db_test.go` が空。データベース層のテストが未実装。
