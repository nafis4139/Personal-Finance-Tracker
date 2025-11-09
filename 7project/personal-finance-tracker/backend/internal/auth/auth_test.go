// backend/internal/auth/auth_test.go
//
// Purpose:
//   Verify password hashing/verification and JWT signing semantics.
// Rationale:
//   These are security-critical primitives; correctness and stability are essential.

package auth_test

import (
	"testing"
	"time"

	"pft/internal/auth"

	"github.com/golang-jwt/jwt/v5"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if err := auth.CheckPassword(hash, "secret123"); err != nil {
		t.Fatalf("expected password to verify, got: %v", err)
	}
	if err := auth.CheckPassword(hash, "wrong"); err == nil {
		t.Fatalf("expected verification to fail for wrong password")
	}
}

func TestSignAndParseClaims(t *testing.T) {
	secret := "testsecret"
	tok, err := auth.Sign(42, secret, time.Hour)
	if err != nil {
		t.Fatalf("sign error: %v", err)
	}
	parsed, err := jwt.Parse(tok, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("token not valid: %v", err)
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("claims type mismatch")
	}
	if uid, ok := claims["uid"].(float64); !ok || int64(uid) != 42 {
		t.Fatalf("unexpected uid claim: %#v", claims["uid"])
	}
}
