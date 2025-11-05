// backend/internal/repo/category.go

package repo

import (
	"context"
	"errors"
	"fmt" // <-- add
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrFKConflict is returned when a delete operation violates a foreign key constraint.
// Used as a sentinel to allow higher layers to map to a 409/Conflict response.
var ErrFKConflict = errors.New("fk_conflict")

// Category is the repository-layer DTO mirroring the categories table.
// Type is expected to be either "income" or "expense".
type Category struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "income" | "expense"
	CreatedAt time.Time `json:"created_at"`
}

// CategoryRepo provides data access for categories via a pgx connection pool.
type CategoryRepo struct{ pool *pgxpool.Pool }

// CategoryRepo constructor bound to the Store's pool.
func (s *Store) CategoryRepo() *CategoryRepo { return &CategoryRepo{pool: s.Pool} }

// List returns all categories for a given user, ordered by id for deterministic output.
func (r *CategoryRepo) List(ctx context.Context, userID int64) ([]Category, error) {
	const q = `SELECT id, user_id, name, type, created_at
	           FROM categories
	           WHERE user_id=$1
	           ORDER BY id`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Type, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Create inserts a new category for the user and returns the inserted row.
// Database constraints (e.g., unique name/type per user) are enforced at the SQL layer.
func (r *CategoryRepo) Create(ctx context.Context, userID int64, name, typ string) (*Category, error) {
	const q = `INSERT INTO categories (user_id, name, type)
	           VALUES ($1,$2,$3)
	           RETURNING id, user_id, name, type, created_at`
	var c Category
	if err := r.pool.QueryRow(ctx, q, userID, name, typ).
		Scan(&c.ID, &c.UserID, &c.Name, &c.Type, &c.CreatedAt); err != nil {
		return nil, err
	}
	return &c, nil
}

// Get fetches a single category by id scoped to the user.
// Returns (nil, nil) when no row is found.
func (r *CategoryRepo) Get(ctx context.Context, userID, id int64) (*Category, error) {
	const q = `SELECT id, user_id, name, type, created_at
	           FROM categories
	           WHERE user_id=$1 AND id=$2`
	var c Category
	err := r.pool.QueryRow(ctx, q, userID, id).Scan(&c.ID, &c.UserID, &c.Name, &c.Type, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// Update modifies name and type for a category owned by the user.
// Returns (nil, nil) if the category is not found (no rows matched).
func (r *CategoryRepo) Update(ctx context.Context, userID, id int64, name, typ string) (*Category, error) {
	const q = `UPDATE categories
	           SET name=$3, type=$4
	           WHERE user_id=$1 AND id=$2
	           RETURNING id, user_id, name, type, created_at`
	var c Category
	err := r.pool.QueryRow(ctx, q, userID, id, name, typ).
		Scan(&c.ID, &c.UserID, &c.Name, &c.Type, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// Delete removes a category by id scoped to the user.
// On foreign key violation (SQLSTATE 23503), returns ErrFKConflict wrapped with the original pg error.
// Returns (false, nil) when no rows were affected.
func (r *CategoryRepo) Delete(ctx context.Context, userID, id int64) (bool, error) {
	const q = `DELETE FROM categories WHERE user_id=$1 AND id=$2`
	tag, err := r.pool.Exec(ctx, q, userID, id)
	if err != nil {
		var pgerr *pgconn.PgError
		// 23503 = foreign_key_violation (likely due to budgets referencing this category).
		if errors.As(err, &pgerr) && pgerr.Code == "23503" { // foreign_key_violation
			// Preserve the sentinel and original error for callers that need details.
			return false, fmt.Errorf("%w: %w", ErrFKConflict, pgerr)
		}
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
