package db

import (
	"database/sql"
	"fmt"

	"github.com/elias-gill/poliplanner2/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

var dbConnection *sql.DB

func InitDB(cfg *config.Config) error {
	// Open database file connection
	db, err := sql.Open("sqlite3", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return fmt.Errorf("enable WAL: %v", err)
	}

	dbConnection = db

	migrateURL := "sqlite3://file:" + cfg.DatabaseURL + "?cache=shared&mode=rwc"

	return runMigrations(cfg.MigrationsDir, migrateURL)
}

func runMigrations(migrationsDir, databaseURL string) error {
	fmt.Println("Running migrations from:", migrationsDir)
	m, err := migrate.New("file://"+migrationsDir, databaseURL)
	if err != nil {
		return fmt.Errorf("create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %v", err)
	}
	return nil
}

func GetConnection() *sql.DB {
	return dbConnection
}

func CloseDB() {
	if dbConnection != nil {
		dbConnection.Close()
	}
}
