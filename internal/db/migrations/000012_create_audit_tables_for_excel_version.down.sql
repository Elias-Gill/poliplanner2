-- NOTE: from here the up and down versions are back again
-- +migrate Down
DROP TABLE sheet_errors;
ALTER TABLE sheet_version DROP COLUMN error_count;
ALTER TABLE sheet_version DROP COLUMN processed_sheets;
ALTER TABLE sheet_version DROP COLUMN total_sheets;
