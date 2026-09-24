-- +migrate Down

CREATE TABLE IF NOT EXISTS auto_sync_excel_check (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_checked_at DATETIME NOT NULL
);

INSERT INTO auto_sync_excel_check (id, last_checked_at)
SELECT 1, COALESCE(last_sync_at, last_search_at)
FROM excel_sync_state
WHERE source_type = 'schedule'
  AND (last_sync_at IS NOT NULL OR last_search_at IS NOT NULL);

DROP TABLE IF EXISTS excel_sync_state;
