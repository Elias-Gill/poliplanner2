-- +migrate Up

-- Audit table: every parse attempt, successful or not.
CREATE TABLE sheet_parse_audit (
    audit_id      INTEGER PRIMARY KEY AUTOINCREMENT,
    version_id    INTEGER,
    source_type   TEXT NOT NULL,
    file_name     TEXT NOT NULL,
    url           TEXT NOT NULL,
    source_date   DATETIME,
    started_at    DATETIME NOT NULL,
    finished_at   DATETIME NOT NULL,
    success       INTEGER NOT NULL,
    error_message TEXT,
    parsed_sheets INTEGER NOT NULL DEFAULT 0
);

-- Preserve existing history as audit rows.
INSERT INTO sheet_parse_audit
    (source_type, file_name, url, source_date, started_at, finished_at, success, error_message, parsed_sheets)
SELECT 'schedule', file_name, url, parsed_at, parsed_at, parsed_at,
       success, COALESCE(error_message, ''), parsed_sheets
FROM sheet_version;

-- Rebuild sheet_version: only successfully parsed versions, with type and source date.
CREATE TABLE sheet_version_new (
    version_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    source_type   TEXT NOT NULL,
    file_name     TEXT NOT NULL,
    url           TEXT NOT NULL,
    source_date   DATETIME,
    parsed_at     DATETIME NOT NULL,
    parsed_sheets INTEGER NOT NULL DEFAULT 0,
    period        INTEGER REFERENCES periodos(id)
);

INSERT INTO sheet_version_new
    (version_id, source_type, file_name, url, source_date, parsed_at, parsed_sheets, period)
SELECT version_id, 'schedule', file_name, url, parsed_at, parsed_at, parsed_sheets, period
FROM sheet_version
WHERE success = 1;

DROP TABLE sheet_version;
ALTER TABLE sheet_version_new RENAME TO sheet_version;

CREATE INDEX idx_sheet_version_name ON sheet_version (source_type, file_name);
CREATE INDEX idx_sheet_version_url  ON sheet_version (source_type, url);
