CREATE TABLE IF NOT EXISTS auto_sync_excel_check (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_checked_at DATETIME NOT NULL
);
