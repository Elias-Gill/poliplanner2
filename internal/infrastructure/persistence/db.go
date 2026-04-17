package persistence

import (
	"database/sql"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/config"
	log "github.com/elias-gill/poliplanner2/logger"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type DbConnection struct {
	db *sql.DB
}

func ConnectDB() (*DbConnection, error) {
	cfg := config.Get()

	log.Info("Initializing database connection", "url", cfg.Database.URL)

	// Open database file connection
	db, err := sql.Open("sqlite3", cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("error opening db: %w", err)
	}
	log.Debug("Database connection established successfully")

	log.Debug("Configuring WAL mode for SQLite")
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("enable WAL: %v", err)
	}
	log.Debug("WAL mode enabled successfully")

	log.Debug("Enabling foreign keys support for SQLite")
	if _, err := db.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %v", err)
	}
	log.Debug("Foreign keys enabled successfully")

	return &DbConnection{db: db}, err
}

func RunMigrations() error {
	cfg := config.Get()

	migrationsDir := cfg.Database.MigrationsDir

	databaseURL := "sqlite3://file:" + cfg.Database.URL + "?cache=shared&mode=rwc"
	log.Debug("Running database migrations", "migrations_dir", cfg.Database.MigrationsDir)

	log.Debug("Creating migration instance", "source", "file://"+migrationsDir)

	m, err := migrate.New("file://"+migrationsDir, databaseURL)
	if err != nil {
		return fmt.Errorf("error creating 'migrate' instance: %v", err)
	}
	defer m.Close()

	// Get current version (before applying)
	prevVersion, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current migration version: %v", err)
	}

	// Apply all pending migrations
	upErr := m.Up()

	// Get version after Up() (even if Up() returned error)
	newVersion, _, verr := m.Version()
	if verr != nil && verr != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version after Up(): %v", verr)
	}

	// Log all applied migrations step by step
	if newVersion > prevVersion {
		for v := prevVersion + 1; v <= newVersion; v++ {
			log.Info("Applied migration", "version", v)
		}
	} else {
		log.Info("No migrations applied - database is up to date")
	}

	// If Up() returned error, report it now
	if upErr != nil && upErr != migrate.ErrNoChange {
		return fmt.Errorf("migration failed after applying version %d: %w", newVersion, upErr)
	}

	return nil
}

func (d *DbConnection) GetConnection() *sql.DB {
	return d.db
}

func (d *DbConnection) CloseDB() {
	if d.db != nil {
		if err := d.db.Close(); err != nil {
			log.Error("Error closing database connection", "error", err)
		} else {
			log.Debug("Database connection closed successfully")
		}
	} else {
		log.Debug("Database connection was already closed or never initialized")
	}
}
