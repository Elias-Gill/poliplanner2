package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/elias-gill/poliplanner2/internal/logger"
)

// Config holds all application configuration
type Config struct {
	// Google API
	GoogleAPIKey string

	// Server
	ServerAddr   string
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

	// Logging
	VerboseLogs bool
}

var cfg *Config = nil

// ================================
// ======== Public API ============
// ================================

func Get() *Config {
	if cfg == nil {
		MustLoad()
	}

	return cfg
}

func SetCustom(c *Config) {
	cfg = c
}

// MustLoad loads configuration from environment variables
func MustLoad() {
	wd, _ := os.Getwd()

	googleAPIKey := getEnv("GOOGLE_API_KEY", "")
	if googleAPIKey == "" {
		log.Warn("Missing Google API Key, web scrapping is disabled")
	}

	cfg = &Config{
		// Google API
		GoogleAPIKey: googleAPIKey,

		// Server
		ServerAddr:   getEnv("SERVER_ADDR", ":8080"),
		ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),

		// Database
		DatabaseURL:   resolvePath(wd, "DATABASE_URL", "poliplanner.db"),
		MigrationsDir: "internal/db/migrations",

		// File paths (resolved from working directory)
		LayoutsDir:   "internal/excelparser/layouts",
		MetadataDir:  "internal/excelparser/metadata",
		DownloadsDir: resolvePath(wd, "DOWNLOADS_DIR", "/tmp/poliplanner/"),

		// Scrapper
		ScrapperTimeout: getEnvAsDuration("SCRAPPER_TIMEOUT", 30*time.Second),
		TargetURL:       getEnv("TARGET_URL", "https://www.pol.una.py/academico/horarios-de-clases-y-examenes/"),

		// Feature flags
		EnableScrapping: getEnvAsBool("ENABLE_SCRAPPING", true),
		EnableDownloads: getEnvAsBool("ENABLE_DOWNLOADS", true),
		VerboseLogs:     getEnvAsBool("VERBOSE_LOGS", false),
	}
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
