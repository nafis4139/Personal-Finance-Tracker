-- backend/migrations/006_category_name_unique.sql
BEGIN;
CREATE UNIQUE INDEX IF NOT EXISTS ux_categories_user_name
  ON categories (user_id, lower(name));
COMMIT;
