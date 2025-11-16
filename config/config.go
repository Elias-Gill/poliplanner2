package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Google API
	GoogleAPIKey string

	// Server
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// Database
	DatabaseURL   string
	MigrationsDir string

	// File paths (absolute paths)
	LayoutsDir   string
	MetadataDir  string
	DownloadsDir string

	// Scrapper
	ScrapperTimeout time.Duration
	TargetURL       string

	// Feature flags
	EnableScrapping bool
	EnableDownloads bool
}

// ================================
// ======== Public API ============
// ================================

// Load loads configuration from environment variables
func Load() *Config {
	wd, _ := os.Getwd()

	cfg := &Config{
		// Google API
		GoogleAPIKey: getEnv("GOOGLE_API_KEY", ""),

		// Server
		Port:         getEnv("PORT", "8080"),
		ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),

		// Database
		DatabaseURL:   resolvePath(wd, "DATABASE_URL", "poliplanner.db"),
		MigrationsDir: resolvePath(wd, "MIGRATIONS_DIR", "db/migrations"),

		// File paths (resolved from working directory)
		LayoutsDir:   resolvePath(wd, "LAYOUTS_DIR", "excelparser/layouts"),
		MetadataDir:  resolvePath(wd, "METADATA_DIR", "excelparser/metadata"),
		DownloadsDir: resolvePath(wd, "DOWNLOADS_DIR", "downloads"),

		// Scrapper
		ScrapperTimeout: getEnvAsDuration("SCRAPPER_TIMEOUT", 30*time.Second),
		TargetURL:       getEnv("TARGET_URL", "https://www.pol.una.py/academico/horarios-de-clases-y-examenes/"),

		// Feature flags
		EnableScrapping: getEnvAsBool("ENABLE_SCRAPPING", true),
		EnableDownloads: getEnvAsBool("ENABLE_DOWNLOADS", true),
	}

	return cfg
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.GoogleAPIKey == "" {
		return fmt.Errorf("GOOGLE_API_KEY is required")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}

// =====================================
// ======== Private methods ============
// =====================================

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// resolvePath returns absolute path:
// - If env var is set: use it (absolute or relative to working directory)
// - Else: join with working directory
func resolvePath(wd, envKey, defaultRel string) string {
	if raw := os.Getenv(envKey); raw != "" {
		if filepath.IsAbs(raw) {
			return raw
		}
		return filepath.Join(wd, raw)
	}
	return filepath.Join(wd, defaultRel)
}
