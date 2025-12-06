FROM ghcr.io/amacneil/dbmate:2.28

COPY ./scripts/migrations/sql /db/migrations

# ワーキングディレクトリを設定（dbmate のデフォルトパスに合わせる）
WORKDIR /db

# データベースURLは実行時に環境変数で指定することを想定
# 例: DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=disable

# マイグレーションを実行
CMD ["dbmate", "up"]
