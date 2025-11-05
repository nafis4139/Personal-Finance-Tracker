// backend/internal/platform/migrate.go

package platform

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations executes all .sql migration files in the provided directory in
// lexicographical order. Applied migrations are tracked in the schema_migrations
// table to ensure idempotency across restarts and deployments.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	entries := []string{}
	// Discover .sql files recursively under dir, skipping subdirectories.
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".sql" {
			entries = append(entries, path)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			// Absence of a migrations directory is treated as a no-op.
			return nil
		}
		return err
	}
	// Sort files to enforce deterministic execution order (e.g., 001_, 002_, ...).
	sort.Strings(entries)

	// Ensure the migrations tracking table exists (safe to run multiple times).
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations(
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT now()
		);
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, f := range entries {
		// Skip files already recorded as applied.
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename=$1)`, filepath.Base(f)).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", f, err)
		}
		if exists {
			continue
		}

		// Read the migration file contents.
		b, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}

		// Execute the migration within a transaction and record it upon success.
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin: %w", err)
		}
		_, err = tx.Exec(ctx, string(b))
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("exec %s: %w", f, err)
		}
		_, err = tx.Exec(ctx, `INSERT INTO schema_migrations(filename) VALUES($1)`, filepath.Base(f))
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", f, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", f, err)
		}
	}
	return nil
}
