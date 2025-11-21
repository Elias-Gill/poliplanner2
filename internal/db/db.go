package db

import (
	"database/sql"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/config"
	log "github.com/elias-gill/poliplanner2/internal/logger"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

var dbConnection *sql.DB

func InitDB(cfg *config.Config) error {
	log.GetLogger().Info("Initializing database connection", "url", cfg.DatabaseURL)

	// Open database file connection
	db, err := sql.Open("sqlite3", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}
	log.GetLogger().Debug("Database connection established successfully")

	log.GetLogger().Debug("Configuring WAL mode for SQLite")
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return fmt.Errorf("enable WAL: %v", err)
	}
	log.GetLogger().Debug("WAL mode enabled successfully")

	dbConnection = db

	migrateURL := "sqlite3://file:" + cfg.DatabaseURL + "?cache=shared&mode=rwc"
	log.GetLogger().Debug("Running database migrations", "migrations_dir", cfg.MigrationsDir)

	return runMigrations(cfg.MigrationsDir, migrateURL)
}

func runMigrations(migrationsDir, databaseURL string) error {
	log.GetLogger().Debug("Creating migration instance", "source", "file://"+migrationsDir)
	m, err := migrate.New("file://"+migrationsDir, databaseURL)
	if err != nil {
		return fmt.Errorf("create migrate instance: %v", err)
	}
	log.GetLogger().Info("Applying database migrations")

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %v", err)
	}

	if err == migrate.ErrNoChange {
		log.GetLogger().Info("Migrations: No changes needed - database is up to date")
	} else {
		log.GetLogger().Info("Migrations applied successfully")
	}
	return nil
}

func GetConnection() *sql.DB {
	return dbConnection
}

func CloseDB() {
	if dbConnection != nil {
		if err := dbConnection.Close(); err != nil {
			log.GetLogger().Error("Error closing database connection", "error", err)
		} else {
			log.GetLogger().Debug("Database connection closed successfully")
		}
	} else {
		log.GetLogger().Debug("Database connection was already closed or never initialized")
	}
}
