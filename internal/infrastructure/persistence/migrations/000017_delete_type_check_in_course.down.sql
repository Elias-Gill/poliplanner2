PRAGMA foreign_keys = OFF;

BEGIN TRANSACTION;

-- 1. Crear tabla volviendo a aplicar CHECK (tipo IN (0, 1))
CREATE TABLE cursos_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    malla INTEGER NOT NULL REFERENCES mallas(id) ON DELETE CASCADE ON UPDATE CASCADE,
    periodo INTEGER NOT NULL REFERENCES periodos(id) ON DELETE CASCADE ON UPDATE CASCADE,
    nombre TEXT NOT NULL,
    seccion VARCHAR(6) NOT NULL,
    turno VARCHAR(10) NOT NULL DEFAULT '',
    tipo INTEGER NOT NULL DEFAULT 0 CHECK (tipo IN (0, 1)),

    comite_presidente TEXT,
    comite_miembro1 TEXT,
    comite_miembro2 TEXT,

    fechas_sabados TEXT,

    UNIQUE (nombre, malla, seccion, periodo, turno)
);

-- 2. Copiar datos mapeando tipos no compatibles (como laboratorio = 3) a 0
INSERT INTO cursos_old (
    id, malla, periodo, nombre, seccion, turno, tipo, 
    comite_presidente, comite_miembro1, comite_miembro2, fechas_sabados
)
SELECT 
    id, malla, periodo, nombre, seccion, turno, 
    CASE WHEN tipo IN (0, 1) THEN tipo ELSE 0 END, 
    comite_presidente, comite_miembro1, comite_miembro2, fechas_sabados 
FROM cursos;

-- 3. Reemplazar la tabla
DROP TABLE cursos;
ALTER TABLE cursos_old RENAME TO cursos;

-- 4. Recrear índices
CREATE INDEX idx_cursos_malla_periodo ON cursos(malla, periodo);

COMMIT;

PRAGMA foreign_keys = ON;
