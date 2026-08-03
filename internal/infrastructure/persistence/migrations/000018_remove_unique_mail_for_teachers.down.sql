-- Desactivar claves foráneas durante la reconstrucción
PRAGMA foreign_keys=OFF;

-- 1. Eliminar los índices creados en la migración UP
DROP INDEX IF EXISTS idx_docentes_nombre_apellido;
DROP INDEX IF EXISTS idx_docentes_correo;

-- 2. Recrear la tabla con el esquema anterior:
--    - correo NOT NULL
--    - Restricción UNIQUE (correo)
CREATE TABLE docentes_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre TEXT NOT NULL,
    apellido TEXT NOT NULL,
    correo TEXT NOT NULL,
    titulo TEXT,
    UNIQUE (correo)
);

-- 3. Restaurar los datos
--    - COALESCE convierte NULLs a texto vacío '' para no violar NOT NULL
--    - INSERT OR IGNORE evita que la migración falle si existen correos duplicados
INSERT OR IGNORE INTO docentes_old (id, titulo, nombre, apellido, correo)
SELECT id, titulo, nombre, apellido, COALESCE(correo, '')
FROM docentes;

-- 4. Eliminar la tabla actual
DROP TABLE docentes;

-- 5. Renombrar la tabla recreada a su nombre original
ALTER TABLE docentes_old RENAME TO docentes;

PRAGMA foreign_keys=ON;
