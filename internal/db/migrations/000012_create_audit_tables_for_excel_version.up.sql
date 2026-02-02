-- +migrate Up
ALTER TABLE sheet_version ADD COLUMN total_sheets INTEGER DEFAULT 0;
ALTER TABLE sheet_version ADD COLUMN processed_sheets INTEGER DEFAULT 0;
ALTER TABLE sheet_version ADD COLUMN error_count INTEGER DEFAULT 0;

CREATE TABLE sheet_errors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version_id INTEGER NOT NULL REFERENCES sheet_version(version_id) ON DELETE CASCADE,
    error_message TEXT NOT NULL,
    created_at DATETIME DEFAULT (datetime('now'))
);

-- New model after insert:
--      * version_id INTEGER
--      * file_name TEXT
--      * file_path TEXT
--      * url TEXT
--      * parsed_at DATETIME
--      * processed_sheets INTEGER
--      * error_count INTEGER
