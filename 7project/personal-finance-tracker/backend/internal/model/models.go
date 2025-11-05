// backend/internal/model/models.go

package model

import "time"

// CategoryType enumerates supported transaction/category kinds.
type CategoryType string

const (
	Income  CategoryType = "income"  // Money coming in
	Expense CategoryType = "expense" // Money going out
)

// User models an account owner in the system.
// Fields map to database columns via `db` struct tags.
type User struct {
	ID           int64     `db:"id"`
	Name         string    `db:"name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

// Category represents a user-scoped classification for transactions.
// Type is constrained by CategoryType (income/expense).
type Category struct {
	ID        int64        `db:"id"`
	UserID    int64        `db:"user_id"`
	Name      string       `db:"name"`
	Type      CategoryType `db:"type"`
	CreatedAt time.Time    `db:"created_at"`
}

// Transaction records a single financial event.
// - CategoryID is nullable to allow uncategorized transactions.
// - Description is optional and stored as a pointer to distinguish empty from null.
// - Type mirrors the category nature (income/expense) at the transaction level.
type Transaction struct {
	ID          int64        `db:"id"`
	UserID      int64        `db:"user_id"`
	CategoryID  *int64       `db:"category_id"`
	Amount      float64      `db:"amount"`
	Type        CategoryType `db:"type"`
	Date        time.Time    `db:"date"`
	Description *string      `db:"description"`
	CreatedAt   time.Time    `db:"created_at"`
}

// Budget defines a monthly spending cap, optionally scoped to a category.
// - PeriodMonth uses the YYYY-MM canonical format.
// - CategoryID is nullable to support an overall (global) budget for the month.
type Budget struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	CategoryID  *int64    `db:"category_id"`
	PeriodMonth string    `db:"period_month"` // YYYY-MM
	LimitAmount float64   `db:"limit_amount"`
	CreatedAt   time.Time `db:"created_at"`
}
