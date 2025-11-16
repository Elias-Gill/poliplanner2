-- +migrate Down
DROP INDEX IF EXISTS idx_careers_name_id;
DROP TABLE IF EXISTS careers;
