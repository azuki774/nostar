# NIP-11: Relay Information Document

## 概要

NIP-11 は、リレーが自身の情報を公開するための標準仕様です。クライアントがリレーの機能を把握し、適切に接続するために必要な情報を JSON 形式で提供します。

## エンドポイント

- **URL**: `/.well-known/nostr.json`
- **メソッド**: GET
- **Content-Type**: application/json

## レスポンス形式

コンフィグファイルにより、レスポンスを定義する。

<details><summary>出力例</summary>

```json
{
  "name": "nostar",
  "description": "A minimal Nostr relay implementation in Go",
  "pubkey": "02...",
  "contact": "admin@example.com",
  "supported_nips": [1, 11, 12, 15, 16, 20, 22],
  "software": "https://github.com/your-org/nostar",
  "version": "0.1.0",
  "relay_countries": ["JP"],
  "language_tags": ["ja"],
  "tags": ["bitcoin", "nostr", "relay"],
  "posting_policy": "https://example.com/posting-policy"
}
```

</details>

## コンフィグ例

```toml
[relay_info]
name = "nostar"
description = "A minimal Nostr relay implementation in Go"
pubkey = "02c7e1b1e9c175ab2d100fce149401e93786935dc7fcb25cef1cac3a4363efc3b5cd"
contact = "admin@example.com"
supported_nips = [1, 11]
relay_countries = ["JP"]
language_tags = ["ja"]
posting_policy = "https://example.com/posting-policy"
```


## 自動で投入されるもの

- `software`
- `version`

## 必須オプション

- `name`: リレーの名前（文字列）
- `description`: リレーの説明文（文字列）
- `pubkey`: リレーの公開鍵（16進文字列、オプション）
- `contact`: 連絡先情報（メールアドレスまたは他の連絡先、オプション）
- `supported_nips`: サポートする NIP の番号リスト（配列）

## オプションの推奨オプション
- `relay_countries`: リレーがホストされている国コードのリスト（ISO 3166-1 alpha-2）
- `language_tags`: サポートする言語タグのリスト（BCP 47）
- `posting_policy`: 投稿ポリシーのURL（文字列）

## 未実装フィールド(TBD)
- `limitation`: リレーの制限事項（オブジェクト）
- `tags`: リレーのトピックを示すタグのリスト

## 実装要件

1. HTTP GET リクエストで上記の JSON を返す
2. 適切な HTTP ヘッダーを設定（Content-Type: application/json）
3. CORS 対応（必要に応じて）
4. 情報は設定ファイルから動的に読み込む
