// backend/internal/handler/category.go

package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"pft/internal/repo"
)

// categoryCreateReq represents the payload for creating a category.
// - Name: human-readable category label
// - Type: constrained to "income" or "expense"
type categoryCreateReq struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
	Type string `json:"type" binding:"required,oneof=income expense"`
}

// categoryUpdateReq mirrors creation fields for updates.
type categoryUpdateReq struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
	Type string `json:"type" binding:"required,oneof=income expense"`
}

// ListCategories returns all categories owned by the authenticated user.
func (api *API) ListCategories(c *gin.Context) {
	userID := MustUserID(c)
	cats, err := api.Repos.CategoryRepo().List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}
	c.JSON(http.StatusOK, cats)
}

// CreateCategory inserts a new category scoped to the authenticated user.
// On unique constraint violation (duplicate name per type), responds with 409.
func (api *API) CreateCategory(c *gin.Context) {
	userID := MustUserID(c)
	var req categoryCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	cat, err := api.Repos.CategoryRepo().Create(c.Request.Context(), userID, req.Name, req.Type)
	if err != nil {
		// Map unique violation (SQLSTATE 23505) to a conflict response.
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "category_exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

// UpdateCategory modifies an existing category by ID.
// Returns 404 if the category does not exist or is not owned by the user.
func (api *API) UpdateCategory(c *gin.Context) {
	userID := MustUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req categoryUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	cat, err := api.Repos.CategoryRepo().Update(c.Request.Context(), userID, id, req.Name, req.Type)
	if err != nil {
		// Handle duplicate name/type combinations as a conflict.
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "category_exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}
	if cat == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, cat)
}

// DeleteCategory removes a category by ID.
// - 409 if a foreign key constraint prevents deletion (e.g., related budgets)
// - 404 if not found
// - 204 on successful deletion
func (api *API) DeleteCategory(c *gin.Context) {
	userID := MustUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	ok, err := api.Repos.CategoryRepo().Delete(c.Request.Context(), userID, id)
	if err != nil {
		// Primary path: sentinel value from repository indicating FK conflict.
		if errors.Is(err, repo.ErrFKConflict) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "category_has_budgets",
				"msg":   "Delete or reassign budgets for this category before deleting it.",
			})
			return
		}
		// Fallback: direct inspection of PostgreSQL FK violation (SQLSTATE 23503).
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23503" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "category_has_budgets",
				"msg":   "Delete or reassign budgets for this category before deleting it.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.Status(http.StatusNoContent)
}
