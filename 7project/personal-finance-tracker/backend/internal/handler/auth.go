// backend/internal/handler/auth.go

package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

// --- Register ---

// registerReq models the expected JSON payload for account creation.
// Validation tags enforce basic constraints on name, email format, and password length.
type registerReq struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// Register creates a new user record and returns a JWT on success.
// - Validates input.
// - Hashes the password with bcrypt.
// - Persists the user; handles unique email violation.
// - Issues a short-lived JWT for immediate authentication.
func (api *API) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}

	// Hash the plaintext password; bcrypt cost 12 balances security and performance.
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}

	// Persist the user; emails are normalized to lowercase for uniqueness.
	// Create is expected to enforce unique(email) at the DB level.
	u, err := api.Repos.UserRepo().Create(c.Request.Context(), req.Name, strings.ToLower(req.Email), string(hashed))
	if err != nil {
		// Handle PostgreSQL unique constraint violation (SQLSTATE 23505).
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "email_in_use"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}

	// Issue a JWT bound to the created user ID with a 24h TTL.
	tok, err := makeToken(api.JWTSecret, u.ID, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": u.ID, "token": tok})
}

// --- Login ---

// loginReq models the expected JSON payload for authentication.
type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login verifies credentials and returns a JWT on success.
// - Looks up user by normalized email.
// - Uses bcrypt constant-time comparison for the password.
// - Returns generic errors to avoid leaking account existence details.
func (api *API) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}

	u, err := api.Repos.UserRepo().GetByEmail(c.Request.Context(), strings.ToLower(req.Email))
	if err != nil || u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}
	// CompareHashAndPassword returns nil on success; any error indicates mismatch or invalid hash.
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}

	// Issue a JWT with a 24h TTL for the authenticated user.
	tok, err := makeToken(api.JWTSecret, u.ID, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": u.ID, "token": tok})
}

// --- Me ---

// Me returns a minimal profile for the authenticated principal.
// Relies on prior middleware to place the user ID in the context.
func (api *API) Me(c *gin.Context) {
	uid := MustUserID(c)
	c.JSON(http.StatusOK, gin.H{"id": uid})
}
