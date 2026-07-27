PRAGMA foreign_keys = OFF;

-- 1. Crear tabla temporal sin el CHECK y con 'turno' incluido
CREATE TABLE cursos_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    malla INTEGER NOT NULL REFERENCES mallas(id) ON DELETE CASCADE ON UPDATE CASCADE,
    periodo INTEGER NOT NULL REFERENCES periodos(id) ON DELETE CASCADE ON UPDATE CASCADE,
    nombre TEXT NOT NULL,
    seccion VARCHAR(6) NOT NULL,
    turno VARCHAR(10) NOT NULL DEFAULT '',
    tipo INTEGER NOT NULL DEFAULT 0,

    comite_presidente TEXT,
    comite_miembro1 TEXT,
    comite_miembro2 TEXT,

    fechas_sabados TEXT,

    UNIQUE (nombre, malla, seccion, periodo, turno)
);

-- 2. Copiar todos los datos existentes
INSERT INTO cursos_new (
    id, malla, periodo, nombre, seccion, turno, tipo, 
    comite_presidente, comite_miembro1, comite_miembro2, fechas_sabados
)
SELECT 
    id, malla, periodo, nombre, seccion, turno, tipo, 
    comite_presidente, comite_miembro1, comite_miembro2, fechas_sabados 
FROM cursos;

-- 3. Reemplazar la tabla vieja por la nueva
DROP TABLE cursos;
ALTER TABLE cursos_new RENAME TO cursos;

-- 4. Recrear índices
CREATE INDEX idx_cursos_malla_periodo ON cursos(malla, periodo);

PRAGMA foreign_keys = ON;
