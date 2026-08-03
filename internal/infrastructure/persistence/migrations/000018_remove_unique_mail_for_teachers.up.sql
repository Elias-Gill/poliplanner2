-- Desactivar claves foráneas durante la reconstrucción
PRAGMA foreign_keys=OFF;

CREATE TABLE docentes_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    titulo TEXT,
    nombre TEXT NOT NULL,
    apellido TEXT NOT NULL,
    correo TEXT
);

-- Migrar los datos existentes
INSERT INTO docentes_new (id, titulo, nombre, apellido, correo)
SELECT id, titulo, nombre, apellido, correo 
FROM docentes;

-- Eliminar la tabla previa
DROP TABLE docentes;

-- Renombrar la tabla nueva a su nombre definitivo
ALTER TABLE docentes_new RENAME TO docentes;

-- Crear índices (no únicos) para optimizar el Upsert
CREATE INDEX idx_docentes_nombre_apellido ON docentes(nombre, apellido);
CREATE INDEX idx_docentes_correo ON docentes(correo);

PRAGMA foreign_keys=ON;
