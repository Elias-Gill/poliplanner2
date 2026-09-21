package config

import (
	"strings"
	"testing"
)

// setBaseEnv prepares a minimal, valid development environment pointing at an
// empty directory so no .env file is loaded.
func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "dev")
	t.Setenv("APP_BASE_DIR", t.TempDir())
	t.Setenv("UPDATE_KEY", "test-key")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("EMAIL_API_KEY", "")
}

func TestLoad_Valid(t *testing.T) {
	setBaseEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Security.UpdateKey != "test-key" {
		t.Errorf("UpdateKey = %q, want %q", cfg.Security.UpdateKey, "test-key")
	}
	if cfg.Server.Addr != ":8080" {
		t.Errorf("Addr = %q, want default %q", cfg.Server.Addr, ":8080")
	}
}

func TestLoad_RequiresUpdateKey(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("UPDATE_KEY", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error when UPDATE_KEY is missing")
	}
	if !strings.Contains(err.Error(), "UPDATE_KEY") {
		t.Errorf("error should mention UPDATE_KEY, got: %v", err)
	}
}

func TestLoad_RejectsMalformedBool(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("VERBOSE_LOGS", "yes")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for a malformed VERBOSE_LOGS value")
	}
	if !strings.Contains(err.Error(), "VERBOSE_LOGS") {
		t.Errorf("error should mention VERBOSE_LOGS, got: %v", err)
	}
}

func TestLoad_RejectsMalformedDuration(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("SCRAPER_TIMEOUT", "30")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for a malformed SCRAPER_TIMEOUT value")
	}
	if !strings.Contains(err.Error(), "SCRAPER_TIMEOUT") {
		t.Errorf("error should mention SCRAPER_TIMEOUT, got: %v", err)
	}
}

func TestLoad_AggregatesErrors(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("UPDATE_KEY", "")
	t.Setenv("VERBOSE_LOGS", "yes")
	t.Setenv("SCRAPER_TIMEOUT", "soon")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error")
	}

	msg := err.Error()
	for _, key := range []string{"UPDATE_KEY", "VERBOSE_LOGS", "SCRAPER_TIMEOUT"} {
		if !strings.Contains(msg, key) {
			t.Errorf("aggregated error should mention %s, got: %v", key, msg)
		}
	}
}
