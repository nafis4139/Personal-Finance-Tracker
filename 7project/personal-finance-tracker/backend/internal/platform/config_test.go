// backend/internal/platform/config_test.go
//
// Purpose:
//   Exercise getenv() defaulting and must() required-variable behavior.

package platform

import (
	"testing"
)

func TestGetenv_DefaultApplied(t *testing.T) {
	t.Setenv("SOME_MISSING_VAR", "")
	if v := getenv("SOME_MISSING_VAR", "fallback"); v != "fallback" {
		t.Fatalf("expected fallback, got %q", v)
	}
}

func TestMust_PanicsWhenMissing(t *testing.T) {
	t.Setenv("ABSENT_VAR", "")
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for missing env")
		}
	}()
	_ = must("ABSENT_VAR")
}

func TestMust_ReturnsValue(t *testing.T) {
	t.Setenv("PRESENT_VAR", "value")
	if v := must("PRESENT_VAR"); v != "value" {
		t.Fatalf("expected value, got %q", v)
	}
}
