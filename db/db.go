package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/elias-gill/poliplanner2/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

var dbConnection *sql.DB

var logger *slog.Logger = slog.Default()

func InitDB(cfg *config.Config) error {
	logger.Info("Initializing database connection", "url", cfg.DatabaseURL)

	// Open database file connection
	db, err := sql.Open("sqlite3", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}
	logger.Debug("Database connection established successfully")

	logger.Debug("Configuring WAL mode for SQLite")
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return fmt.Errorf("enable WAL: %v", err)
	}
	logger.Debug("WAL mode enabled successfully")

	dbConnection = db

	migrateURL := "sqlite3://file:" + cfg.DatabaseURL + "?cache=shared&mode=rwc"
	logger.Debug("Running database migrations", "migrations_dir", cfg.MigrationsDir)

	return runMigrations(cfg.MigrationsDir, migrateURL)
}

func runMigrations(migrationsDir, databaseURL string) error {
	logger.Debug("Creating migration instance", "source", "file://"+migrationsDir)
	m, err := migrate.New("file://"+migrationsDir, databaseURL)
	if err != nil {
		return fmt.Errorf("create migrate instance: %v", err)
	}
	logger.Info("Applying database migrations...")

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %v", err)
	}

	if err == migrate.ErrNoChange {
		logger.Info("Migrations: No changes needed - database is up to date")
	} else {
		logger.Info("Migrations applied successfully")
	}
	return nil
}

func GetConnection() *sql.DB {
	return dbConnection
}

func CloseDB() {
	if dbConnection != nil {
		if err := dbConnection.Close(); err != nil {
			logger.Error("Error closing database connection", "error", err)
		} else {
			logger.Debug("Database connection closed successfully")
		}
	} else {
		logger.Debug("Database connection was already closed or never initialized")
	}
}
