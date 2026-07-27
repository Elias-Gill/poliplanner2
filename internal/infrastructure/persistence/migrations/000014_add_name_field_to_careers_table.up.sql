-- +migrate Up

ALTER TABLE carreras ADD COLUMN nombre TEXT default '';
