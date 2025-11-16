-- +migrate Down
DROP INDEX IF EXISTS idx_subjects_name;
DROP TABLE IF EXISTS subjects;
