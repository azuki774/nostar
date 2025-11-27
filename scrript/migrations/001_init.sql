CREATE TABLE events (
  -- Nostr イベント本体
  id          CHAR(64) PRIMARY KEY,  -- イベントID（SHA-256 hex）
  pubkey      CHAR(64) NOT NULL,     -- 発行者のpubkey（hex）
  created_at  BIGINT NOT NULL,       -- NIP-01のUNIX時刻
  kind        INTEGER NOT NULL,      -- イベント種別
  tags        JSONB   NOT NULL,      -- [["e","..."],["p","..."], ...]
  content     TEXT    NOT NULL,      -- 本文
  sig         CHAR(128) NOT NULL,    -- 署名（hex）

  -- リレー側メタデータ
  received_at TIMESTAMPTZ NOT NULL DEFAULT now(), -- リレーが受信した時刻
  deleted     BOOLEAN NOT NULL DEFAULT FALSE,     -- NIP-09 等で無効化済みか
  hidden      BOOLEAN NOT NULL DEFAULT FALSE,     -- モデレーション用
  replaced_by CHAR(64) NULL                       -- 置き換えられた先のID（オプション）
);

-- 作成時刻で範囲検索する用（since/until）
CREATE INDEX idx_events_created_at
  ON events (created_at);

-- pubkey + created_at で「あるユーザのタイムライン」を引く
CREATE INDEX idx_events_pubkey_created_at
  ON events (pubkey, created_at DESC);

-- kind ごとの抽出（プロフィールだけ、コンタクトリストだけ等）
CREATE INDEX idx_events_kind_created_at
  ON events (kind, created_at DESC);

-- tags の #e/#p 検索用（JSONB GIN）
CREATE INDEX idx_events_tags_gin
  ON events USING GIN (tags);
