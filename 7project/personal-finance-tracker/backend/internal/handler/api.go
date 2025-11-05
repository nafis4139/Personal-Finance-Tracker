// backend/internal/handler/api.go
package handler

import (
	"net/http"
	"strconv"

	"pft/internal/repo"

	"github.com/gin-gonic/gin"
)

// API groups HTTP handlers with their required dependencies.
// - Repos: data access layer for persistence operations
// - JWTSecret: symmetric key used by middleware/handlers for JWT validation or signing
type API struct {
	Repos     *repo.Store
	JWTSecret string
}

// New constructs an API instance with injected dependencies.
// Keeps all handler methods stateless by passing required services via this struct.
func New(repos *repo.Store, jwtSecret string) *API {
	return &API{Repos: repos, JWTSecret: jwtSecret}
}

// --- existing handlers already present elsewhere ---

// Healthz exposes a simple health endpoint intended for readiness/liveness probes.
// Returns HTTP 200 with a minimal JSON payload when the process is responsive.
func (api *API) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Register, Login, Me ... (present in other files)

// --- helpers ---

// MustUserID extracts the authenticated user's ID from the Gin context.
// Assumes a prior authentication middleware has stored the claim under key "uid".
// Supported underlying types: int64, int, string (base-10). Any other type results in panic.
// Panics if the key is missing; this is deliberate to fail fast when middleware is misconfigured.
// For a non-panicking variant, consider returning (int64, bool) instead.
func MustUserID(c *gin.Context) int64 {
	uidVal, ok := c.Get("uid")
	if !ok {
		// If the middleware uses a different key name (e.g., "user_id"),
		// adapt the lookup accordingly.
		panic("uid missing in context")
	}
	switch v := uidVal.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	default:
		panic("bad uid type")
	}
}
