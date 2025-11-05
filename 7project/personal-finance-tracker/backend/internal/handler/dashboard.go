// backend/internal/handler/dashboard.go

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MonthSummary returns an aggregate view for a given month.
// Expects query parameter "month" in YYYY-MM format.
// Responds with 400 if the month is missing, 500 on repository errors, and 200 with the summary payload on success.
func (api *API) MonthSummary(c *gin.Context) {
	userID := MustUserID(c)
	month := c.Query("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month_required"})
		return
	}
	out, err := api.Repos.DashboardRepo().Summary(c.Request.Context(), userID, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}
	c.JSON(http.StatusOK, out)
}
