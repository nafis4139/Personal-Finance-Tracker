// backend/internal/platform/config.go

package platform

import (
	"fmt"
	"os"
)

// Config holds application configuration derived from environment variables.
// Fields:
//   - Port: HTTP listen port (e.g., "8080")
//   - DB_DSN: database connection string
//   - JWTSecret: HMAC secret for JWT signing/verification
type Config struct {
	Port      string
	DB_DSN    string
	JWTSecret string
}

// Load constructs a Config by reading environment variables.
// Defaults:
//   - PORT defaults to "8080" if unset.
//
// Required:
//   - DB_DSN must be set or the process panics.
//   - JWT_SECRET must be set or the process panics.
func Load() Config {
	return Config{
		Port:      getenv("PORT", "8080"),
		DB_DSN:    must("DB_DSN"),
		JWTSecret: must("JWT_SECRET"),
	}
}

// getenv returns the value of environment variable k, or default d if empty.
// Keeps a single-line style for brevity while avoiding extraneous branching.
func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

// must returns the value of required environment variable k.
// Panics with a descriptive message if k is not present or empty.
func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Sprintf("missing env %s", k))
	}
	return v
}
