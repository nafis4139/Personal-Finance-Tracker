// backend/internal/handler/jwt.go

package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig holds configuration for JWT-based authentication.
// Currently includes the HMAC secret used to sign/verify tokens.
type AuthConfig struct {
	JWTSecret string
}

// JWTMiddleware validates a Bearer JWT from the Authorization header.
// Flow:
//  1. Require "Authorization: Bearer <token>" header.
//  2. Parse and verify the token using HMAC (HS256).
//  3. Extract the "uid" claim and store it in the context for downstream handlers.
//  4. Abort with 401 on any validation failure.
func JWTMiddleware(cfg AuthConfig) gin.HandlerFunc {
	secret := []byte(cfg.JWTSecret)

	return func(c *gin.Context) {
		// Expect a Bearer token in the Authorization header.
		ah := c.GetHeader("Authorization")
		if !strings.HasPrefix(ah, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing bearer token"})
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(ah, "Bearer "))

		// Parse and validate the JWT signature and algorithm.
		tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// Enforce HMAC-based signing methods; reject unexpected algorithms.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return secret, nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		// Extract claims as a generic map; expect "uid" to be present.
		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid claims"})
			return
		}
		// JSON numbers decode to float64; convert to int64 for application use.
		uidF, ok := claims["uid"].(float64)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "uid missing"})
			return
		}
		// Store the user ID in the Gin context for later retrieval.
		c.Set("uid", int64(uidF))
		c.Next()
	}
}

// makeToken issues a signed JWT containing:
//   - "uid": application user ID
//   - "exp": expiration timestamp (Unix seconds) derived from ttl
//
// Uses HS256 with the provided secret.
func makeToken(secret string, uid int64, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"uid": uid,
		"exp": time.Now().Add(ttl).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}
