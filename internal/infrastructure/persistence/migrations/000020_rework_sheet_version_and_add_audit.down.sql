-- +migrate Down

DROP INDEX IF EXISTS idx_sheet_version_url;
DROP INDEX IF EXISTS idx_sheet_version_name;

CREATE TABLE sheet_version_old (
    version_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    file_name     TEXT NOT NULL,
    url           TEXT NOT NULL,
    parsed_at     DATETIME NOT NULL DEFAULT (datetime('now')),
    success       INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    parsed_sheets INTEGER NOT NULL DEFAULT 0,
    period        INTEGER REFERENCES periodos(id)
);

INSERT INTO sheet_version_old
    (version_id, file_name, url, parsed_at, success, error_message, parsed_sheets, period)
SELECT version_id, file_name, url, parsed_at, 1, NULL, parsed_sheets, period
FROM sheet_version;

DROP TABLE sheet_version;
ALTER TABLE sheet_version_old RENAME TO sheet_version;

DROP TABLE IF EXISTS sheet_parse_audit;
