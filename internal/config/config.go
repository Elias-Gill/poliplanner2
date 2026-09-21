package config

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elias-gill/poliplanner2/logger"
)

const (
	appName = "PoliPlanner"
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
	Name string
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
	// Dir is the directory where the active log file and its backup are
	// written. It is created on demand.
	Dir string
	// MaxSizeBytes rotates the log file once it would grow past this size.
	MaxSizeBytes int64
	// RotateAfter rotates the log file once it is older than this duration.
	RotateAfter time.Duration
}

type SecurityConfig struct {
	UpdateKey  string
	SecureHTTP bool
}

type EmailConfig struct {
	APIKey string
}

// ================================
//        Secret redaction        =
// ================================

// redact hides a secret while distinguishing it from an empty value.
func redact(secret string) string {
	if secret == "" {
		return ""
	}
	return "***"
}

// LogValue implements slog.LogValuer so a Config can be logged without leaking
// secrets in its nested groups.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app", c.App),
		slog.Any("server", c.Server),
		slog.Any("database", c.Database),
		slog.Any("paths", c.Paths),
		slog.Any("excel", c.Excel),
		slog.Any("logging", c.Logging),
		slog.Any("security", c.Security),
		slog.Any("email", c.Email),
	)
}

func (s SecurityConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("update_key", redact(s.UpdateKey)),
		slog.Bool("secure_http", s.SecureHTTP),
	)
}

func (e ExcelConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("google_api_key", redact(e.GoogleAPIKey)),
		slog.Duration("scraper_timeout", e.ScraperTimeout),
	)
}

func (e EmailConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("api_key", redact(e.APIKey)),
	)
}

var (
	once    sync.Once // To ensure thread safety
	cfg     *Config
	loadErr error
)

// ================================
//         Public API             =
// ================================

// Init loads the configuration exactly once and caches it. It is safe to call
// multiple times; only the first call performs any work. It should run before
// Get, typically from the application entry point.
func Init() (*Config, error) {
	once.Do(func() {
		cfg, loadErr = load()
	})
	return cfg, loadErr
}

// Get returns the cached configuration. If the configuration failed to load,
// Get panics with the underlying error, since the application cannot run
// without a valid configuration.
func Get() *Config {
	c, err := Init()
	if err != nil {
		panic(fmt.Sprintf("config: initialization failed: %v", err))
	}
	return c
}

// Load builds a fresh configuration without touching the cache. It is intended
// for tests and tooling that need an isolated configuration.
func Load() (*Config, error) {
	return load()
}

// ================================
// =        Internal load         =
// ================================

func load() (*Config, error) {
	env := Environment(os.Getenv("APP_ENV"))
	if env != EnvDev && env != EnvProd {
		env = EnvDev
	}

	l := &loader{}

	baseDir, err := resolveBaseDir(env)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve base dir: %w", err)
	}

	// Never load .env files if the production flag is set
	if env == EnvDev {
		loadDotenv(filepath.Join(baseDir, ".env"))
	}

	updateKey := l.required("UPDATE_KEY")
	googleAPIKey := l.string("GOOGLE_API_KEY", "")
	emailAPIKey := l.string("EMAIL_API_KEY", "")

	// Secure http on production. Unsecure for dev to avoid local network problems
	secureHTTPDefault := env == EnvProd

	// Verbose logs disabled on production
	verboseLogsDefault := env == EnvDev

	cfg := &Config{
		App: AppData{
			Name: appName,
		},

		Server: ServerConfig{
			Addr: l.string("SERVER_ADDR", ":8080"),
			Env:  env,
		},

		Database: DatabaseConfig{
			URL:           l.path(baseDir, "DATABASE_URL", "poliplanner.db"),
			MigrationsDir: l.path(baseDir, "", "internal/infrastructure/persistence/migrations"),
		},

		Paths: PathsConfig{
			BaseDir:                baseDir,
			ExcelParsingLayoutsDir: l.path(baseDir, "", "internal/infrastructure/parser/layouts"),
			MetadataDir:            l.path(baseDir, "", "internal/service/metadata/data"),
			DownloadsDir:           l.path(baseDir, "DOWNLOADS_DIR", "tmp/poliplanner"),
			TemplatesDir:           l.path(baseDir, "", "internal/render/html/templates"),
			AssetsDir:              l.path(baseDir, "", "web"),
		},

		Excel: ExcelConfig{
			GoogleAPIKey:   googleAPIKey,
			ScraperTimeout: l.duration("SCRAPER_TIMEOUT", 30*time.Second),
		},

		Logging: LoggingConfig{
			Verbose:      l.bool("VERBOSE_LOGS", verboseLogsDefault),
			Dir:          l.path(baseDir, "LOG_DIR", "logs"),
			MaxSizeBytes: int64(l.integer("LOG_MAX_SIZE_MB", 10)) * 1024 * 1024,
			RotateAfter:  l.duration("LOG_ROTATE_INTERVAL", 15*24*time.Hour),
		},

		Security: SecurityConfig{
			UpdateKey:  updateKey,
			SecureHTTP: l.bool("SECURE_HTTP", secureHTTPDefault),
		},

		Email: EmailConfig{
			APIKey: emailAPIKey,
		},
	}

	if len(l.errs) > 0 {
		return nil, errors.Join(l.errs...)
	}

	return cfg, nil
}

// ================================
// =         Helpers              =
// ================================

func resolveBaseDir(env Environment) (string, error) {
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

	// Production must be explicit: the working directory is not a reliable
	// source for the application layout, as it depends on how the process is
	// launched.
	if env == EnvProd {
		return "", fmt.Errorf("APP_BASE_DIR must be set when APP_ENV=%s", EnvProd)
	}

	// In development the process often runs from a package subdirectory (for
	// example `go test ./...`), so we walk up to the module root to keep the
	// relative Paths (layouts, metadata, templates, ...) pointing at the
	// repository.
	if root, ok := findModuleRoot(wd); ok {
		return root, nil
	}

	return wd, nil
}

// findModuleRoot walks up from dir and returns the first directory containing a
// go.mod file.
func findModuleRoot(dir string) (string, bool) {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// ================================
//         Env helpers            =
// ================================

// loader reads environment variables and accumulates validation errors so the
// whole configuration can be reported at once instead of failing silently on
// the first malformed value.
type loader struct {
	errs []error
}

// string returns the variable value or the provided default when unset.
func (l *loader) string(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// required records an error when the variable is unset.
func (l *loader) required(key string) string {
	value := os.Getenv(key)
	if value == "" {
		l.errs = append(l.errs, fmt.Errorf("missing required environment variable %s", key))
	}
	return value
}

// bool parses a boolean variable, recording an error when it is malformed.
func (l *loader) bool(key string, defaultValue bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("invalid boolean value for %s: %q", key, raw))
		return defaultValue
	}
	return value
}

// integer parses an integer variable, recording an error when it is malformed.
func (l *loader) integer(key string, defaultValue int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("invalid integer value for %s: %q", key, raw))
		return defaultValue
	}
	return value
}

// duration parses a duration variable, recording an error when it is
// malformed.
func (l *loader) duration(key string, defaultValue time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("invalid duration value for %s: %q", key, raw))
		return defaultValue
	}
	return value
}

// path resolves a file or directory relative to baseDir. When envKey is not
// empty and the variable is defined it takes precedence: absolute values are
// used as is and relative ones are joined to baseDir.
func (l *loader) path(baseDir, envKey, defaultRel string) string {
	if envKey != "" {
		if raw := os.Getenv(envKey); raw != "" {
			if filepath.IsAbs(raw) {
				return raw
			}
			return filepath.Join(baseDir, raw)
		}
	}
	return filepath.Join(baseDir, defaultRel)
}

func loadDotenv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignore empty lines and full-line comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Strip 'export ' prefix if present
		if after, ok := strings.CutPrefix(line, "export "); ok {
			line = strings.TrimSpace(after)
		}

		before, after, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key := strings.TrimSpace(before)
		if key == "" {
			continue
		}

		// Do not overwrite existing system environment variables
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		raw := strings.TrimSpace(after)

		// Handle quoted values ("val" or 'val')
		if (strings.HasPrefix(raw, `"`) && strings.HasSuffix(raw, `"`)) ||
			(strings.HasPrefix(raw, "'") && strings.HasSuffix(raw, "'")) {
			if len(raw) >= 2 {
				raw = raw[1 : len(raw)-1]
			}
		} else {
			// Unquoted values: strip inline comments only if preceded by space (" #")
			// Keeps values with '#' untouched (e.g., DB_PASS=secret#123)
			if comment := strings.Index(raw, " #"); comment != -1 {
				raw = strings.TrimSpace(raw[:comment])
			}
		}

		os.Setenv(key, raw)
	}

	// Check for scanner errors after loop finishes
	if err := scanner.Err(); err != nil {
		logger.Warn("Error reading .env file", "error", err.Error())
	}
}
