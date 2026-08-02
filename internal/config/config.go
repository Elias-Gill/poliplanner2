package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	appName    = "PoliPlanner"
	appVersion = "1.0.0"
)

const (
	EnvDev  Environment = "dev"
	EnvProd Environment = "prod"
)

// ================================
// =           Config             =
// ================================

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Paths    PathsConfig
	Excel    ExcelConfig
	Logging  LoggingConfig
	Security SecurityConfig
	Email    EmailConfig
	App      AppData
}

// ================================
// =       Sub-config types       =
// ================================

type Environment string

type ServerConfig struct {
	Addr string
	Env  Environment
}

type AppData struct {
	Name    string
	Version string
}

type DatabaseConfig struct {
	URL           string
	MigrationsDir string
}

type PathsConfig struct {
	BaseDir                string
	ExcelParsingLayoutsDir string
	MetadataDir            string
	DownloadsDir           string
	TemplatesDir           string
	AssetsDir              string
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

type EmailConfig struct {
	APIKey string
}

// ================================
// =         Load .env file       =
// ================================

func loadDotenv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq == -1 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		if key == "" {
			continue
		}
		if os.Getenv(key) != "" {
			continue
		}
		raw := strings.TrimSpace(line[eq+1:])
		if comment := strings.IndexByte(raw, '#'); comment != -1 {
			raw = strings.TrimSpace(raw[:comment])
		}
		os.Setenv(key, raw)
	}
}

// ================================
// =        Resolve base dir      =
// ================================

var (
	cfg  *Config
	err  error
	once sync.Once // To ensure thread safety
)

// ================================
// =         Public API           =
// ================================

// Get returns loaded config or nil if failed.
func Get() *Config {
	once.Do(func() {
		cfg, err = load()
	})
	return cfg
}

// Err returns initialization error if any.
func Err() error {
	once.Do(func() {
		cfg, err = load()
	})
	return err
}

// Load builds configuration from environment variables.
func Load() (*Config, error) {
	return load()
}

// ================================
// =        Internal load         =
// ================================

func load() (*Config, error) {
	loadDotenv(".env")

	env := Environment(getEnv("APP_ENV", "dev"))
	if env != EnvDev && env != EnvProd {
		env = EnvDev
	}

	baseDir, err := resolveBaseDir()
	if err != nil {
		return nil, fmt.Errorf("cannot resolve base dir: %w", err)
	}

	googleAPIKey := getEnv("GOOGLE_API_KEY", "")

	emailAPIKey := getEnv("EMAIL_API_KEY", "")

	updateKey := getEnv("UPDATE_KEY", "")
	if updateKey == "" {
		return nil, fmt.Errorf("missing UPDATE_KEY")
	}

	secureHTTPDefault := env == EnvProd // secure on production

	verboseLogsDefault := env == EnvDev // verbose on development

	cfg := &Config{
		App: AppData{
			Name:    appName,
			Version: appVersion,
		},

		Server: ServerConfig{
			Addr: getEnv("SERVER_ADDR", ":8080"),
			Env:  env,
		},

		Database: DatabaseConfig{
			URL:           resolveOrDefaultPath(baseDir, "DATABASE_URL", "poliplanner.db"),
			MigrationsDir: filepath.Join(baseDir, "internal", "infrastructure", "persistence", "migrations"),
		},

		Paths: PathsConfig{
			BaseDir:                baseDir,
			ExcelParsingLayoutsDir: filepath.Join(baseDir, "internal", "infrastructure", "parser", "layouts"),
			MetadataDir:            filepath.Join(baseDir, "internal", "service", "metadata", "data"),
			DownloadsDir:           resolveOrDefaultPath(baseDir, "DOWNLOADS_DIR", filepath.Join("tmp", "poliplanner")),
			TemplatesDir:           filepath.Join(baseDir, "web", "templates"),
			AssetsDir:              filepath.Join(baseDir, "web", "static"),
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

		Email: EmailConfig{
			APIKey: emailAPIKey,
		},
	}

	return cfg, nil
}

// ================================
// =         Helpers              =
// ================================

func resolveBaseDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if raw := os.Getenv("APP_BASE_DIR"); raw != "" {
		if filepath.IsAbs(raw) {
			return raw, nil
		}
		return filepath.Join(wd, raw), nil
	}

	return wd, nil
}

// ================================
// =         Env helpers          =
// ================================

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
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

func resolveOrDefaultPath(baseDir, envKey, defaultRel string) string {
	if raw := os.Getenv(envKey); raw != "" {
		if filepath.IsAbs(raw) {
			return raw
		}
		return filepath.Join(baseDir, raw)
	}
	return filepath.Join(baseDir, defaultRel)
}
