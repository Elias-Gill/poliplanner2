package config

import (
	"fmt"
	"os"
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
	DatabaseURL string

	// File paths
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
	return &Config{
		// Google API
		GoogleAPIKey: getEnv("GOOGLE_API_KEY", ""),

		// Server
		Port:         getEnv("PORT", "8080"),
		ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", ""),

		// File paths
		LayoutsDir:   getEnv("LAYOUTS_DIR", "layouts"),
		MetadataDir:  getEnv("METADATA_DIR", "metadata"),
		DownloadsDir: getEnv("DOWNLOADS_DIR", "downloads"),

		// Scrapper
		ScrapperTimeout: getEnvAsDuration("SCRAPPER_TIMEOUT", 30*time.Second),
		TargetURL:       getEnv("TARGET_URL", "https://www.pol.una.py/academico/horarios-de-clases-y-examenes/"),

		// Feature flags
		EnableScrapping: getEnvAsBool("ENABLE_SCRAPPING", true),
		EnableDownloads: getEnvAsBool("ENABLE_DOWNLOADS", true),
	}
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
