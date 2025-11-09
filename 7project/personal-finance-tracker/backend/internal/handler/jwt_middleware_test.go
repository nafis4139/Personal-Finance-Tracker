// backend/internal/handler/jwt_middleware_test.go
//
// Purpose:
//   Validate the JWT middleware's request gating behavior for valid and invalid tokens.
// Method:
//   Use a tiny Gin app with the middleware and a dummy protected handler.

package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pft/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func makeToken(t *testing.T, secret string, uid int64) string {
	t.Helper()
	claims := jwt.MapClaims{
		"uid": uid,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}

func TestJWTMiddleware_AllowsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "s3cr3t"
	r := gin.New()
	r.Use(handler.JWTMiddleware(handler.AuthConfig{JWTSecret: secret}))
	r.GET("/protected", func(c *gin.Context) {
		uidAny, _ := c.Get("uid")
		c.JSON(200, gin.H{"uid": uidAny})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+makeToken(t, secret, 99))
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if got := int(body["uid"].(float64)); got != 99 {
		t.Fatalf("expected uid=99, got %v", body["uid"])
	}
}

func TestJWTMiddleware_RejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "s3cr3t"
	r := gin.New()
	r.Use(handler.JWTMiddleware(handler.AuthConfig{JWTSecret: secret}))
	r.GET("/protected", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil) // no header
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401, got %d (body: %s)", w.Code, w.Body.String())
	}
}
