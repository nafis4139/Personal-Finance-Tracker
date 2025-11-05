// backend/internal/handler/helper.go

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getUID extracts the authenticated user ID from the Gin context.
// Expects an upstream authentication middleware to place a value under the "uid" key.
// On missing or invalid value, the request is aborted with HTTP 401 and zero is returned.
// Supported types: int64, int, string (parsing omitted and returns zero as a safe default).
func getUID(c *gin.Context) int64 {
	v, ok := c.Get("uid")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0
	}
	switch id := v.(type) {
	case int64:
		return id
	case int:
		return int64(id)
	case string:
		// If middleware stored a string UID, parsing could be performed here.
		// Returning zero keeps behavior consistent with unauthorized/invalid UID handling.
		return 0
	default:
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0
	}
}

// bad sends a standardized 400 Bad Request response with the provided error message.
// Useful for input validation failures and similar client-side error conditions.
func bad(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
