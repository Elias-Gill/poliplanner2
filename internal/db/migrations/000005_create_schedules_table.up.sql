-- +migrate Up
CREATE TABLE schedules (
    schedule_id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    user_id INTEGER NOT NULL,
    schedule_description TEXT NOT NULL,
    schedule_sheet_version INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (schedule_sheet_version) REFERENCES sheet_version(version_id) ON DELETE CASCADE
);
