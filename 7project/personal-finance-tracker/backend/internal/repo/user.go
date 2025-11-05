// backend/internal/repo/user.go

package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User represents a row from the users table.
// JSON tags hide the password hash from API responses.
type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`          // excluded from JSON output
	CreatedAt    time.Time `json:"created_at"` // server-set timestamp
}

// UserRepo provides basic access methods for the users table.
type UserRepo struct{ pool *pgxpool.Pool }

// UserRepo getter on Store, mirroring the pattern used by other repositories.
func (s *Store) UserRepo() *UserRepo { return &UserRepo{pool: s.Pool} }

// Create inserts a new user with a previously computed password hash.
// Returns the inserted row, including generated ID and timestamps.
func (r *UserRepo) Create(ctx context.Context, name, email, passwordHash string) (*User, error) {
	const q = `
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, name, email, password_hash, created_at`
	var u User
	if err := r.pool.QueryRow(ctx, q, name, email, passwordHash).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByEmail returns a user identified by email.
// On no match, returns (nil, nil) rather than an error.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	const q = `
SELECT id, name, email, password_hash, created_at
FROM users
WHERE email = $1`
	var u User
	if err := r.pool.QueryRow(ctx, q, email).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// GetByID returns a user by primary key.
// On no match, returns (nil, nil) rather than an error.
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*User, error) {
	const q = `
SELECT id, name, email, password_hash, created_at
FROM users
WHERE id = $1`
	var u User
	if err := r.pool.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
