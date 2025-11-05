-- +goose Up
-- +migrate Up
BEGIN;

-- Drop the existing FK (currently ON DELETE SET NULL)
ALTER TABLE budgets
  DROP CONSTRAINT IF EXISTS budgets_category_id_fkey;

-- Re-add it with ON DELETE RESTRICT (the default)
ALTER TABLE budgets
  ADD CONSTRAINT budgets_category_id_fkey
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;

COMMIT;

-- +goose Down
-- +migrate Down
BEGIN;

-- Roll back to SET NULL if you ever need to
ALTER TABLE budgets
  DROP CONSTRAINT IF EXISTS budgets_category_id_fkey;

ALTER TABLE budgets
  ADD CONSTRAINT budgets_category_id_fkey
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;

COMMIT;