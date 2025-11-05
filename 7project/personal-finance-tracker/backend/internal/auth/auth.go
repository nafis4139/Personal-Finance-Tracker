// backend/internal/auth/auth.go
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims defines the custom JWT payload.
// - UserID: application-specific subject identifier
// - RegisteredClaims: standard fields such as exp, iat, nbf
type Claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

// HashPassword returns a bcrypt hash for the given plaintext password.
// bcrypt.DefaultCost is used to balance security and performance.
// The returned string is the encoded hash suitable for storage.
func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword compares a bcrypt hash with a candidate plaintext password.
// Returns nil if the password matches the hash; otherwise returns an error
// from bcrypt indicating mismatch or invalid hash format.
func CheckPassword(hash, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

// Sign creates a JWT token string for the provided userID using an HMAC-SHA256
// signing method and the supplied secret. The token includes an expiration
// time set to now + ttl. Additional standard claims can be added as needed
// (e.g., Issuer, Subject, IssuedAt) by extending RegisteredClaims.
func Sign(userID int64, secret string, ttl time.Duration) (string, error) {
	c := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString([]byte(secret))
}
