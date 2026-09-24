-- +migrate Up

CREATE TABLE IF NOT EXISTS excel_sync_state (
    source_type    TEXT PRIMARY KEY,

    -- Last web search (discovery) for this source type. Set even when nothing
    -- new was found; gates the search interval.
    last_search_at DATETIME,

    -- Last successful persistence of new sources for this source type. Only set
    -- when there was something pending to sync.
    last_sync_at   DATETIME
);

INSERT INTO excel_sync_state (source_type, last_search_at, last_sync_at)
SELECT 'schedule', last_checked_at, last_checked_at
FROM auto_sync_excel_check
WHERE id = 1;

DROP TABLE IF EXISTS auto_sync_excel_check;
