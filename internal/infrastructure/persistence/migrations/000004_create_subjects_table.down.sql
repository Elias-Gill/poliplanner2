-- +migrate Down
DROP INDEX IF EXISTS idx_subjects_name_career;
DROP TABLE IF EXISTS subjects;
