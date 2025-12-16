package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/elias-gill/poliplanner2/internal/logger"
)

// ================================
// ========= Environment ==========
// ================================

type Environment string

const (
	EnvDev  Environment = "dev"
	EnvProd Environment = "prod"
)

// ================================
// =========== Config =============
// ================================

// Config holds all application configuration.
// It is loaded once at startup and treated as immutable afterwards.
type Config struct {
	Environment Environment

	Server   ServerConfig
	Database DatabaseConfig
	Paths    PathsConfig
	Excel    ExcelConfig
	Logging  LoggingConfig
	Security SecurityConfig
}

// ================================
// ========= Sub-configs ===========
// ================================

type ServerConfig struct {
	Addr string
}

type DatabaseConfig struct {
	URL           string
	MigrationsDir string
}

type PathsConfig struct {
	BaseDir                string
	ExcelParsingLayoutsDir string
	SubjectsMetadataDir    string
	DownloadsDir           string
}

type ExcelConfig struct {
	GoogleAPIKey   string
	ScraperTimeout time.Duration
}

type LoggingConfig struct {
	Verbose bool
}

type SecurityConfig struct {
	UpdateKey  string
	SecureHTTP bool
}

// ================================
// ========= Global state ==========
// ================================

var cfg *Config = nil

// ================================
// ========= Public API ============
// ================================

// Get returns the loaded configuration.
func Get() *Config {
	if cfg == nil {
		MustLoad()
	}
	return cfg
}

// SetCustom allows injecting a custom config (useful for tests).
func SetCustom(c *Config) {
	cfg = c
}

// MustLoad loads configuration from environment variables.
// The application exits if critical configuration is missing.
func MustLoad() {
	env := Environment(getEnv("APP_ENV", "dev"))
	if env != EnvDev && env != EnvProd {
		log.Warn("Unknown APP_ENV value, defaulting to dev", "value", env)
		env = EnvDev
	}

	// Resolve base directory (defaults to executable directory)
	baseDir := resolveBaseDir()

	googleAPIKey := getEnv("GOOGLE_API_KEY", "")
	if googleAPIKey == "" {
		log.Warn("Missing Google API key, excel scrapping will be disabled")
	}

	updateKey := getEnv("UPDATE_KEY", "")
	if updateKey == "" {
		log.Error("Missing UPDATE_KEY, refusing to start")
		os.Exit(1)
	}

	secureHTTPDefault := env == EnvProd // true on production
	verboseLogsDefault := env == EnvDev // false on production

	cfg = &Config{
		Environment: env,

		Server: ServerConfig{
			Addr: getEnv("SERVER_ADDR", ":8080"),
		},

		Database: DatabaseConfig{
			URL:           resolvePath(baseDir, "DATABASE_URL", "poliplanner.db"),
			MigrationsDir: filepath.Join(baseDir, "internal/db/migrations"),
		},

		Paths: PathsConfig{
			BaseDir:                baseDir,
			ExcelParsingLayoutsDir: filepath.Join(baseDir, "internal/excelparser/layouts"),
			SubjectsMetadataDir:    filepath.Join(baseDir, "internal/excelparser/metadata"),
			DownloadsDir:           resolvePath(baseDir, "DOWNLOADS_DIR", "tmp/poliplanner"),
		},

		Excel: ExcelConfig{
			GoogleAPIKey:   googleAPIKey,
			ScraperTimeout: getEnvAsDuration("SCRAPER_TIMEOUT", 30*time.Second),
		},

		Logging: LoggingConfig{
			Verbose: getEnvAsBool("VERBOSE_LOGS", verboseLogsDefault),
		},

		Security: SecurityConfig{
			UpdateKey:  updateKey,
			SecureHTTP: getEnvAsBool("SECURE_HTTP", secureHTTPDefault),
		},
	}
}

// ================================
// ========= Helpers ==============
// ================================

// resolveBaseDir determines the application base directory.
// Priority:
// 1. APP_BASE_DIR env var
// 2. Current working directory
func resolveBaseDir() string {
	wd, _ := os.Getwd()

	if raw := os.Getenv("APP_BASE_DIR"); raw != "" {
		if filepath.IsAbs(raw) {
			return raw
		} else {
			return filepath.Join(wd, raw)
		}
	}

	return wd
}

// ================================
// ========= Env helpers ==========
// ================================

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

// resolvePath returns an absolute path.
// If the env var is set, it is used (absolute or relative to baseDir).
// Otherwise, the default relative path is joined with baseDir.
func resolvePath(baseDir, envKey, defaultRel string) string {
	if raw := os.Getenv(envKey); raw != "" {
		if filepath.IsAbs(raw) {
			return raw
		}
		return filepath.Join(baseDir, raw)
	}
	return filepath.Join(baseDir, defaultRel)
}
