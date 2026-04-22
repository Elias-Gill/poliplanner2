-- +migrate Up
-- SQLite permite crear la FK aunque la tabla referenciada no exista aún
CREATE TABLE career (
    career_id INTEGER PRIMARY KEY AUTOINCREMENT,
    career_code TEXT NOT NULL UNIQUE
);

CREATE INDEX idx_career_code ON career(career_code);

CREATE TABLE career_version (
    career_version_id INTEGER PRIMARY KEY AUTOINCREMENT,
    career_id INTEGER NOT NULL,
    sheet_version_id INTEGER NOT NULL,
    FOREIGN KEY (career_id) REFERENCES career(career_id) ON DELETE CASCADE,
    FOREIGN KEY (sheet_version_id) REFERENCES sheet_version(version_id) ON DELETE CASCADE,
-- Una carrera solo puede tener una versión por sheet
    UNIQUE (career_id, sheet_version_id)
);

CREATE INDEX idx_career_version_sheet ON career_version(sheet_version_id);

