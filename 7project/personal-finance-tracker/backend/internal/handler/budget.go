// backend/internal/handler/budget.go

package handler

import (
	"net/http"
	"strconv"

	"pft/internal/repo"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// budgetCreateReq models the payload for creating or updating a budget.
// - CategoryID: optional category scoping (nil means a global/monthly budget)
// - PeriodMonth: target period in YYYY-MM format
// - LimitAmount: allowed spending limit for the period/category
type budgetCreateReq struct {
	CategoryID  *int64  `json:"category_id"`                     // nullable
	PeriodMonth string  `json:"period_month" binding:"required"` // YYYY-MM
	LimitAmount float64 `json:"limit_amount" binding:"required"`
}

// Reuse the same shape for updates; all fields are handled similarly.
type budgetUpdateReq = budgetCreateReq

// ListBudgets retrieves all budgets for the authenticated user limited to a single month.
// Requires the "month" query parameter in YYYY-MM format.
func (api *API) ListBudgets(c *gin.Context) {
	userID := MustUserID(c)
	month := c.Query("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month_required"})
		return
	}
	out, err := api.Repos.BudgetRepo().ListByMonth(c.Request.Context(), userID, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// CreateBudget inserts a new budget row for the authenticated user.
// Enforces uniqueness at the database level to prevent duplicate
// (user, category_id, period_month) entries.
func (api *API) CreateBudget(c *gin.Context) {
	userID := MustUserID(c)
	var req budgetCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	b := &repo.Budget{
		UserID:      userID,
		CategoryID:  req.CategoryID,
		PeriodMonth: req.PeriodMonth,
		LimitAmount: req.LimitAmount,
	}
	out, err := api.Repos.BudgetRepo().Create(c.Request.Context(), b)
	if err != nil {
		// Handle unique-constraint violations (SQLSTATE 23505),
		// which indicate a budget already exists for the given key.
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "budget_exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}
	c.JSON(http.StatusCreated, out)
}

// UpdateBudget modifies an existing budget owned by the authenticated user.
// Budget ID is provided as a path parameter; fields mirror creation semantics.
func (api *API) UpdateBudget(c *gin.Context) {
	userID := MustUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req budgetUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	b := &repo.Budget{
		CategoryID:  req.CategoryID,
		PeriodMonth: req.PeriodMonth,
		LimitAmount: req.LimitAmount,
	}
	out, err := api.Repos.BudgetRepo().Update(c.Request.Context(), userID, id, b)
	if err != nil {
		// Preserve unique constraint handling for conflicting keys.
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "budget_exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// DeleteBudget removes a budget by ID for the authenticated user.
// Returns 204 on success, 404 if the budget does not exist or is not owned by the user.
func (api *API) DeleteBudget(c *gin.Context) {
	userID := MustUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	ok, err := api.Repos.BudgetRepo().Delete(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.Status(http.StatusNoContent)
}
