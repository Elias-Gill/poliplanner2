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

func InitDB() (*DbConnection, error) {
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

	migrateURL := "sqlite3://file:" + cfg.Database.URL + "?cache=shared&mode=rwc"
	log.Debug("Running database migrations", "migrations_dir", cfg.Database.MigrationsDir)

	err = runMigrations(cfg.Database.MigrationsDir, migrateURL)

	return &DbConnection{db: db}, err
}

func runMigrations(migrationsDir, databaseURL string) error {
	log.Debug("Creating migration instance", "source", "file://"+migrationsDir)
	m, err := migrate.New("file://"+migrationsDir, databaseURL)
	if err != nil {
		return fmt.Errorf("error creating 'migrate' instance: %v", err)
	}
	defer m.Close()

	for {
		if err := m.Steps(1); err != nil {
			if err == migrate.ErrNoChange {
				log.Info("No more migrations to apply, database is up to date")
				break
			}
			return fmt.Errorf("migration failed: %v", err)
		}

		// Log migration info
		version, _, verr := m.Version()
		if verr != nil {
			return fmt.Errorf("failed to get migration version: %v", verr)
		}
		log.Info("Applied migration", "version", version)
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
