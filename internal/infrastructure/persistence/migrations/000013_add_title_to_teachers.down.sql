-- +migrate Down

ALTER TABLE docentes DROP COLUMN titulo;

ALTER TABLE docentes ADD search_key TEXT;
CREATE INDEX idx_docentes_search_key ON docentes(search_key);
