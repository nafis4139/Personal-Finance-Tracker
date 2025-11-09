// Run with: PG_TEST_DSN="postgres://app:app@localhost:5432/app?sslmode=disable" go test ./integration -v
package integration

import (
	"context"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"pft/internal/handler"
	"pft/internal/platform"
	"pft/internal/repo"
)

func setup(t *testing.T) (*gin.Engine, func()) {
	t.Helper()
	dsn := os.Getenv("PG_TEST_DSN")
	if dsn == "" {
		t.Skip("PG_TEST_DSN not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	pcfg, _ := pgxpool.ParseConfig(dsn)
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		t.Fatal(err)
	}

	// migrate
	if err := platform.RunMigrations(ctx, pool, "../../migrations"); err != nil {
		t.Fatal(err)
	}

	api := handler.New(repo.New(pool), "testsecret")
	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/api/register", api.Register)
	r.POST("/api/login", api.Login)
	r.GET("/api/healthz", api.Healthz)

	// clean tables between tests? You can add TRUNCATE here if needed

	return r, func() { pool.Close() }
}

func TestRegisterLogin(t *testing.T) {
	r, _ := setup(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(`{"name":"t","email":"t@e.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 && w.Code != 409 { // allow reruns
		t.Fatalf("register got %d: %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"email":"t@e.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("login got %d: %s", w.Code, w.Body.String())
	}
}
