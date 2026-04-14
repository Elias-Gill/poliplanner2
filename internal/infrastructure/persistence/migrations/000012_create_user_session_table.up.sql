-- +migrate Up

CREATE TABLE IF NOT EXISTS data_migrations (
    key TEXT PRIMARY KEY,
    executed_at DATETIME NOT NULL DEFAULT (datetime('now')),
    notes TEXT
);
