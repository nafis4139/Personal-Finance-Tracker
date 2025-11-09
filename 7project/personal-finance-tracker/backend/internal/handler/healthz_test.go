// backend/internal/handler/healthz_test.go
//
// Purpose:
//   This unit test verifies that the public health endpoint `/api/healthz`
//   responds with HTTP 200 OK and a minimal JSON payload `{ "ok": true }`.
// Rationale:
//   The health check is used by platforms (e.g., Render, Docker healthcheck)
//   to assess process liveness/readiness. A fast, dependency-free handler is
//   expected to be stable and deterministic. Testing it in isolation helps
//   ensure operability signals remain correct over time.
//
// Method:
//   The test constructs a lightweight Gin router in test mode, registers the
//   health handler, and invokes it using the standard library's `httptest`
//   utilities. No external dependencies (database, network) are involved,
//   keeping the test hermetic and fast.
//
// Acceptance Criteria:
//   - Status code MUST be 200.
//   - Body MUST be parseable JSON containing a boolean field `ok` set to true.

package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pft/internal/handler"

	"github.com/gin-gonic/gin"
)

func TestHealthz_OK(t *testing.T) {
	// Arrange: minimize logging/noise during tests.
	gin.SetMode(gin.TestMode)

	// Construct the API; Healthz does not depend on repositories,
	// so a nil repo store is acceptable here.
	api := handler.New(nil, "testsecret")

	router := gin.New()
	router.GET("/api/healthz", api.Healthz)

	// Act: perform a GET request against the registered route.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	router.ServeHTTP(rec, req)

	// Assert: HTTP status is 200 OK.
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	// Assert: response body is JSON with { "ok": true }.
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v (raw: %s)", err, rec.Body.String())
	}
	ok, _ := body["ok"].(bool)
	if !ok {
		t.Fatalf(`expected body to contain {"ok": true}, got: %s`, rec.Body.String())
	}
}
