package config

import (
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestSecretLogValuesHideSecrets(t *testing.T) {
	cases := map[string]string{
		"security": SecurityConfig{UpdateKey: "super-secret", SecureHTTP: true}.LogValue().String(),
		"excel":    ExcelConfig{GoogleAPIKey: "google-secret"}.LogValue().String(),
		"email":    EmailConfig{APIKey: "email-secret"}.LogValue().String(),
	}

	for name, out := range cases {
		t.Run(name, func(t *testing.T) {
			if strings.Contains(out, "super-secret") ||
				strings.Contains(out, "google-secret") ||
				strings.Contains(out, "email-secret") {
				t.Errorf("LogValue leaked a secret: %s", out)
			}
		})
	}
}

func TestLoad_LoggingDefaultsInDev(t *testing.T) {
	setBaseEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// File logging is off by default in development.
	if cfg.Logging.Dir != "" {
		t.Errorf("Logging.Dir = %q, want empty in dev", cfg.Logging.Dir)
	}
	if cfg.Logging.MaxSizeBytes != 10*1024*1024 {
		t.Errorf("Logging.MaxSizeBytes = %d, want %d", cfg.Logging.MaxSizeBytes, 10*1024*1024)
	}
	if cfg.Logging.RotateAfter != 15*24*time.Hour {
		t.Errorf("Logging.RotateAfter = %s, want %s", cfg.Logging.RotateAfter, 15*24*time.Hour)
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Logging.Format = %q, want text", cfg.Logging.Format)
	}
	// Dev is verbose by default, so the level defaults to debug.
	if cfg.Logging.Level != slog.LevelDebug {
		t.Errorf("Logging.Level = %v, want debug", cfg.Logging.Level)
	}
}

func TestLoad_LoggingDefaultsInProd(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("APP_BASE_DIR", t.TempDir())
	t.Setenv("UPDATE_KEY", "test-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// File logging is on by default in production.
	wantDir := filepath.Join(cfg.Paths.BaseDir, "logs")
	if cfg.Logging.Dir != wantDir {
		t.Errorf("Logging.Dir = %q, want %q", cfg.Logging.Dir, wantDir)
	}
	if cfg.Logging.Level != slog.LevelInfo {
		t.Errorf("Logging.Level = %v, want info", cfg.Logging.Level)
	}
}

func TestLoad_LogDirOverride(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("LOG_TO_FILE", "true")
	t.Setenv("LOG_DIR", "/var/log/poliplanner")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Logging.Dir != "/var/log/poliplanner" {
		t.Errorf("Logging.Dir = %q, want %q", cfg.Logging.Dir, "/var/log/poliplanner")
	}
}

func TestLoad_LogLevelAndFormat(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("LOG_FORMAT", "json")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Logging.Level != slog.LevelWarn {
		t.Errorf("Logging.Level = %v, want warn", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format = %q, want json", cfg.Logging.Format)
	}
}

func TestLoad_RejectsInvalidLogValues(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("LOG_LEVEL", "loud")
	t.Setenv("LOG_FORMAT", "xml")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for invalid log level and format")
	}
	for _, want := range []string{"LOG_LEVEL", "LOG_FORMAT"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %s, got: %v", want, err)
		}
	}
}

func TestLoad_ProdRequiresBaseDir(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("APP_BASE_DIR", "")
	t.Setenv("UPDATE_KEY", "test-key")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error when APP_BASE_DIR is missing in production")
	}
	if !strings.Contains(err.Error(), "APP_BASE_DIR") {
		t.Errorf("error should mention APP_BASE_DIR, got: %v", err)
	}
}

func TestLoad_ProdWithBaseDir(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("APP_BASE_DIR", t.TempDir())
	t.Setenv("UPDATE_KEY", "test-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Env != EnvProd {
		t.Errorf("Env = %q, want %q", cfg.Server.Env, EnvProd)
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
