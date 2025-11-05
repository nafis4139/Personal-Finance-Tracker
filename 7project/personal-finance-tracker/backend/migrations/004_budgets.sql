-- 004_budgets.sql
CREATE TABLE IF NOT EXISTS budgets (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id   BIGINT NULL REFERENCES categories(id) ON DELETE SET NULL,
    period_month  TEXT NOT NULL CHECK (period_month ~ '^[0-9]{4}-[0-9]{2}$'), -- YYYY-MM
    limit_amount  NUMERIC(12,2) NOT NULL CHECK (limit_amount >= 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_budgets_user_month_cat
  ON budgets(user_id, period_month, COALESCE(category_id, -1));
