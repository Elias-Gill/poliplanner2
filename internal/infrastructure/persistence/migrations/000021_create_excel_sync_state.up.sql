-- +migrate Up

CREATE TABLE IF NOT EXISTS excel_sync_state (
    source_type    TEXT PRIMARY KEY,
    last_search_at DATETIME,
    last_sync_at   DATETIME
);

INSERT INTO excel_sync_state (source_type, last_search_at, last_sync_at)
SELECT 'schedule', last_checked_at, last_checked_at
FROM auto_sync_excel_check
WHERE id = 1;

DROP TABLE IF EXISTS auto_sync_excel_check;
