// backend/internal/model/budget.go
package model

// Budget represents a monthly spending limit for a specific category.
// Fields:
//   - CategoryID: identifier of the category the budget applies to
//   - Month: billing period in YYYY-MM format
//   - Limit: maximum allowed spending for the period (currency units)
type Budget struct {
	CategoryID int64   `json:"category_id"`
	Month      string  `json:"month"`        // YYYY-MM
	Limit      float64 `json:"limit_amount"` // monetary amount cap for the month
}
