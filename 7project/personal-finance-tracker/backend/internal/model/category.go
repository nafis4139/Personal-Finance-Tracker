// backend/internal/model/category.go
package model

// Category defines a classification for financial transactions.
// Fields:
//   - ID: unique identifier for the category
//   - Name: descriptive label (e.g., "Groceries", "Salary")
//   - Type: indicates whether the category represents "income" or "expense"
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // income | expense
}
