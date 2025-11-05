// backend/internal/repo/db.go

package repo

import "github.com/jackc/pgx/v5/pgxpool"

// Store wraps a shared pgx connection pool.
// Repository constructors are exposed as methods on Store,
// enabling access patterns like api.Repos.UserRepo().
type Store struct {
	Pool *pgxpool.Pool
}

// New constructs a Store bound to the provided connection pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{Pool: pool}
}
