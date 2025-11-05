-- 003_transactions.sql
CREATE TABLE IF NOT EXISTS transactions (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id  BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    amount       NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    type         TEXT NOT NULL CHECK (type IN ('income','expense')),
    date         DATE NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tx_user_date      ON transactions(user_id, date);
CREATE INDEX IF NOT EXISTS idx_tx_user_category  ON transactions(user_id, category_id);
CREATE INDEX IF NOT EXISTS idx_tx_user_type      ON transactions(user_id, type);
