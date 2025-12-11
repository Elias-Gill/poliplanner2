-- +migrate Up
-- SQLite permite crear la FK aunque la tabla referenciada no exista aún
CREATE TABLE careers (
    career_id INTEGER PRIMARY KEY AUTOINCREMENT,
    career_code TEXT NOT NULL,
    sheet_version_id INTEGER,
    FOREIGN KEY (sheet_version_id) REFERENCES sheet_version(version_id) ON DELETE CASCADE
);

CREATE INDEX idx_careers_code_id ON careers(career_code, career_id);
