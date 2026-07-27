-- Desactivar Foreign Keys temporalmente
PRAGMA foreign_keys = OFF;

-- Eliminar tablas nuevas creadas
DROP TABLE IF EXISTS enfasis_materia;
DROP TABLE IF EXISTS enfasis;
DROP TABLE IF EXISTS mallas;
DROP TABLE IF EXISTS planes;

-- Recrear la tabla 'mallas' original (sin la columna 'plan')
CREATE TABLE mallas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    carrera INTEGER NOT NULL REFERENCES carreras(id) ON DELETE CASCADE ON UPDATE CASCADE,
    asignatura INTEGER NOT NULL REFERENCES asignaturas(id) ON DELETE CASCADE ON UPDATE CASCADE,
    semestre INTEGER NOT NULL DEFAULT 0,
    nivel INTEGER NOT NULL DEFAULT 0,

    -- Restricción única original
    UNIQUE (carrera, asignatura)
);

CREATE INDEX IF NOT EXISTS idx_mallas_carrera_semestre ON mallas(carrera);

PRAGMA foreign_keys = ON;
PRAGMA foreign_key_check;
