-- +migrate Up

ALTER TABLE docentes ADD titulo VARCHAR(20);

DROP INDEX idx_docentes_search_key;
ALTER TABLE docentes DROP COLUMN search_key;
