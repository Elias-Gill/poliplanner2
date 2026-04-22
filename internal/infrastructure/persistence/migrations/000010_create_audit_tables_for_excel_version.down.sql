-- NOTE: from here the up and down versions are back again
-- +migrate Down
ALTER TABLE sheet_version DROP COLUMN success;
ALTER TABLE sheet_version DROP COLUMN error_message;
