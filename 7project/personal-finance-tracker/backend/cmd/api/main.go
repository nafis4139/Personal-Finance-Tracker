// cmd/api/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"pft/internal/handler"
	"pft/internal/platform"
	"pft/internal/repo"
)

func main() {
	// Load application configuration from environment or defaults.
	// Expected fields include DB_DSN, JWTSecret, and Port.
	cfg := platform.Load()

	// --- Database: establish a pooled connection (pgxpool) ---
	// Use a timeout to avoid hanging during startup if the DB is unreachable.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse the DSN into a pgx pool configuration.
	pcfg, err := pgxpool.ParseConfig(cfg.DB_DSN)
	if err != nil {
		log.Fatalf("pgx parse config: %v", err)
	}

	// Create the connection pool using the parsed configuration.
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		log.Fatalf("pgxpool new: %v", err)
	}

	// Verify connectivity before proceeding.
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	defer pool.Close()

	// --- Database: run schema migrations ---
	// Expects migration files to be available at /migrations (e.g., mounted in container).
	if err := platform.RunMigrations(ctx, pool, "/migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// --- Repository and HTTP handler wiring ---
	// repo.Store abstracts access to the database; handler.API uses it plus JWT secret.
	store := repo.New(pool)                  // Repository layer wrapping *pgxpool.Pool
	api := handler.New(store, cfg.JWTSecret) // HTTP API with dependencies injected

	// --- HTTP server setup (Gin) ---
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery()) // Request logging and panic recovery
	_ = r.SetTrustedProxies(nil)        // Disable proxy trust to rely on direct client IPs

	// ===== Public endpoints (no authentication required) =====
	r.GET("/api/healthz", api.Healthz)    // Liveness/readiness probe
	r.POST("/api/register", api.Register) // Account creation
	r.POST("/api/login", api.Login)       // JWT issuance

	// ===== Authenticated endpoints (JWT required) =====
	// Configure JWT middleware with the shared secret.
	authMw := handler.JWTMiddleware(handler.AuthConfig{JWTSecret: cfg.JWTSecret})
	auth := r.Group("/api", authMw)

	// Me
	auth.GET("/me", api.Me) // Return authenticated user profile

	// Categories
	auth.GET("/categories", api.ListCategories)        // List categories for the user
	auth.POST("/categories", api.CreateCategory)       // Create a category
	auth.PUT("/categories/:id", api.UpdateCategory)    // Update a category by ID
	auth.DELETE("/categories/:id", api.DeleteCategory) // Delete a category by ID

	// Transactions
	auth.GET("/transactions", api.ListTransactions)         // List transactions with optional filters
	auth.POST("/transactions", api.CreateTransaction)       // Create a transaction
	auth.PUT("/transactions/:id", api.UpdateTransaction)    // Update a transaction by ID
	auth.DELETE("/transactions/:id", api.DeleteTransaction) // Delete a transaction by ID

	// Budgets
	auth.GET("/budgets", api.ListBudgets)         // List budgets, supports ?month=YYYY-MM
	auth.POST("/budgets", api.CreateBudget)       // Create a budget
	auth.PUT("/budgets/:id", api.UpdateBudget)    // Update a budget by ID
	auth.DELETE("/budgets/:id", api.DeleteBudget) // Delete a budget by ID

	// Dashboard
	auth.GET("/dashboard/summary", api.MonthSummary) // Aggregate monthly summary, ?month=YYYY-MM

	// ---- HTTP server configuration and startup ----
	srv := &http.Server{
		Addr:    ":" + cfg.Port, // Bind to configured port on all interfaces
		Handler: r,              // Gin engine as the HTTP handler
	}

	// Start the server in a separate goroutine to allow graceful shutdown handling.
	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// ---- Graceful shutdown ----
	// Listen for termination signals (SIGINT/SIGTERM) to shut down cleanly.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	// Allow in-flight requests to complete within a timeout.
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	// Attempt a graceful server shutdown.
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	log.Println("server stopped cleanly")
}
