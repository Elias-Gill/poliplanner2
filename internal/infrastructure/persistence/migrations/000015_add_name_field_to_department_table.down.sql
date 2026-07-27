-- +migrate Down

ALTER TABLE departamentos DROP COLUMN nombre;
