-- backend/migrations/005_constraints_indexes.sql
BEGIN;

-- 1) Make transactions.category_id nullable so ON DELETE SET NULL is possible
ALTER TABLE transactions
  ALTER COLUMN category_id DROP NOT NULL;

-- 2) Drop old FKs if they exist (names may differ locally; these IF EXISTS guards are safe)
ALTER TABLE categories   DROP CONSTRAINT IF EXISTS categories_user_id_fkey;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_user_id_fkey;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_category_id_fkey;
ALTER TABLE budgets      DROP CONSTRAINT IF EXISTS budgets_user_id_fkey;

-- 3) Re-add FKs with desired delete behavior
ALTER TABLE categories
  ADD CONSTRAINT categories_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_category_id_fkey
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;

ALTER TABLE budgets
  ADD CONSTRAINT budgets_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- 4) Helpful indexes
CREATE INDEX IF NOT EXISTS idx_tx_user_date
  ON transactions (user_id, date);

CREATE INDEX IF NOT EXISTS idx_tx_user_category
  ON transactions (user_id, category_id);

CREATE INDEX IF NOT EXISTS idx_bud_user_month
  ON budgets (user_id, period_month);

-- Optional (recommended): prevent duplicate category names per user (case-insensitive)
CREATE UNIQUE INDEX IF NOT EXISTS ux_categories_user_name
  ON categories (user_id, lower(name));

-- Optional (recommended): type safety (only 'income' or 'expense')
ALTER TABLE categories
  DROP CONSTRAINT IF EXISTS chk_categories_type;
ALTER TABLE categories
  ADD CONSTRAINT chk_categories_type CHECK (type IN ('income','expense'));

ALTER TABLE transactions
  DROP CONSTRAINT IF EXISTS chk_transactions_type;
ALTER TABLE transactions
  ADD CONSTRAINT chk_transactions_type CHECK (type IN ('income','expense'));

COMMIT;
