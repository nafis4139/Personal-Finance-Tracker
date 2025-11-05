// backend/internal/repo/dashboard.go

package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MonthSummary aggregates totals for a given month.
// - Month: string period identifier in YYYY-MM
// - IncomeTotal/ExpenseTotal: summed amounts by type
type MonthSummary struct {
	Month        string  `json:"month"` // YYYY-MM
	IncomeTotal  float64 `json:"income_total"`
	ExpenseTotal float64 `json:"expense_total"`
}

// DashboardRepo provides read-only aggregation queries for dashboard views.
type DashboardRepo struct{ pool *pgxpool.Pool }

// DashboardRepo accessor bound to the Store's connection pool.
func (s *Store) DashboardRepo() *DashboardRepo { return &DashboardRepo{pool: s.Pool} }

// Summary returns income and expense totals for a specific month.
// The month parameter should be in YYYY-MM format.
// Computes an inclusive date range [first day, last instant of month] and
// executes a single SQL query using conditional aggregation.
func (r *DashboardRepo) Summary(ctx context.Context, userID int64, month string) (*MonthSummary, error) {
	// Derive the first day of the month; ignore parse error since month is validated upstream.
	first, _ := time.Parse("2006-01", month)
	// Compute the last instant of the month: start of next month minus 1ns.
	last := first.AddDate(0, 1, 0).Add(-time.Nanosecond)

	const q = `
SELECT
	COALESCE(SUM(CASE WHEN type='income' THEN amount END),0) AS income_total,
	COALESCE(SUM(CASE WHEN type='expense' THEN amount END),0) AS expense_total
FROM transactions
WHERE user_id=$1 AND date >= $2 AND date <= $3
`
	var m MonthSummary
	m.Month = month
	if err := r.pool.QueryRow(ctx, q, userID, first, last).Scan(&m.IncomeTotal, &m.ExpenseTotal); err != nil {
		return nil, err
	}
	return &m, nil
}
