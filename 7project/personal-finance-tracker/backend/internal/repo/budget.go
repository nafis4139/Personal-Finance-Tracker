// backend/internal/repo/budget.go
package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Budget is the repository-layer DTO mirroring the budgets table.
// CategoryID is nullable to support global (uncategorized) monthly budgets.
type Budget struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	CategoryID  *int64    `json:"category_id"`  // nullable
	PeriodMonth string    `json:"period_month"` // YYYY-MM
	LimitAmount float64   `json:"limit_amount"`
	CreatedAt   time.Time `json:"created_at"`
}

// BudgetRepo provides CRUD operations for budgets using a pgx connection pool.
type BudgetRepo struct{ pool *pgxpool.Pool }

// BudgetRepo returns a BudgetRepo bound to the Store's pool.
func (s *Store) BudgetRepo() *BudgetRepo { return &BudgetRepo{pool: s.Pool} }

// ListByMonth fetches all budgets for a given user and YYYY-MM period.
// Results are ordered by id for deterministic client rendering.
func (r *BudgetRepo) ListByMonth(ctx context.Context, userID int64, month string) ([]Budget, error) {
	const q = `SELECT id, user_id, category_id, period_month, limit_amount, created_at
	           FROM budgets
	           WHERE user_id=$1 AND period_month=$2
	           ORDER BY id`
	rows, err := r.pool.Query(ctx, q, userID, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Budget
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.ID, &b.UserID, &b.CategoryID, &b.PeriodMonth, &b.LimitAmount, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// Create inserts a new budget and returns the inserted row, including timestamps.
func (r *BudgetRepo) Create(ctx context.Context, b *Budget) (*Budget, error) {
	const q = `INSERT INTO budgets (user_id, category_id, period_month, limit_amount)
	           VALUES ($1,$2,$3,$4)
	           RETURNING id, user_id, category_id, period_month, limit_amount, created_at`
	var out Budget
	if err := r.pool.QueryRow(ctx, q, b.UserID, b.CategoryID, b.PeriodMonth, b.LimitAmount).
		Scan(&out.ID, &out.UserID, &out.CategoryID, &out.PeriodMonth, &out.LimitAmount, &out.CreatedAt); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update modifies an existing budget owned by userID and returns the updated row.
// Matching on both user_id and id ensures tenant isolation.
func (r *BudgetRepo) Update(ctx context.Context, userID, id int64, b *Budget) (*Budget, error) {
	const q = `UPDATE budgets
	           SET category_id=$3, period_month=$4, limit_amount=$5
	           WHERE user_id=$1 AND id=$2
	           RETURNING id, user_id, category_id, period_month, limit_amount, created_at`
	var out Budget
	err := r.pool.QueryRow(ctx, q, userID, id, b.CategoryID, b.PeriodMonth, b.LimitAmount).
		Scan(&out.ID, &out.UserID, &out.CategoryID, &out.PeriodMonth, &out.LimitAmount, &out.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a budget by id scoped to userID.
// Returns true when a row was deleted, false if nothing matched.
func (r *BudgetRepo) Delete(ctx context.Context, userID, id int64) (bool, error) {
	const q = `DELETE FROM budgets WHERE user_id=$1 AND id=$2`
	ct, err := r.pool.Exec(ctx, q, userID, id)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}
