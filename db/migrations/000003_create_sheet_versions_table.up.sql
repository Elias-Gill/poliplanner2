-- +migrate Up
CREATE TABLE sheet_version (
    version_id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_name TEXT NOT NULL,
    url TEXT NOT NULL,
    parsed_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
