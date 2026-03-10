-- +migrate Up
ALTER TABLE sheet_version ADD COLUMN processed_sheets INTEGER DEFAULT 0;
ALTER TABLE sheet_version ADD COLUMN succeeded_sheets INTEGER DEFAULT 0;

CREATE TABLE sheet_errors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version_id INTEGER NOT NULL REFERENCES sheet_version(version_id) ON DELETE CASCADE,
    error_message TEXT NOT NULL,
    created_at DATETIME DEFAULT (datetime('now'))
);
