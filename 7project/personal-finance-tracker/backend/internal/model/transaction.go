// backend/internal/model/transaction.go

package model

// Transaction represents a single financial record as exposed via JSON models.
// Fields:
//   - ID: unique transaction identifier
//   - CategoryID: optional category reference; omitted in JSON when nil
//   - Amount: monetary value of the transaction
//   - Direction: flow indicator; expected values are "in" or "out"
//   - Note: optional human-readable description; omitted when empty
//   - OccurredAt: transaction date in YYYY-MM-DD format
type Transaction struct {
	ID         int64   `json:"id"`
	CategoryID *int64  `json:"category_id,omitempty"`
	Amount     float64 `json:"amount"`
	Direction  string  `json:"direction"` // in | out
	Note       string  `json:"note,omitempty"`
	OccurredAt string  `json:"occurred_at"` // YYYY-MM-DD
}
