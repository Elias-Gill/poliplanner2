-- +migrate Up
ALTER TABLE sheet_version ADD COLUMN success INTEGER NOT NULL DEFAULT 0;
ALTER TABLE sheet_version ADD COLUMN error_message TEXT;
