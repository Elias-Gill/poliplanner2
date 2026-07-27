-- +migrate Up

ALTER TABLE departamentos ADD COLUMN nombre TEXT default '';
