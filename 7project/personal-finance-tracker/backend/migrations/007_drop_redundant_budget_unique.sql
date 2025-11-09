-- backend/migrations/007_drop_redundant_budget_unique.sql
BEGIN;
ALTER TABLE budgets
  DROP CONSTRAINT IF EXISTS budgets_user_id_category_id_period_month_key;
COMMIT;
