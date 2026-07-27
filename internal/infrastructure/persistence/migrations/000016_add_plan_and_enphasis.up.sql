-- ==============================================================================
-- MIGRACIÓN: Reestructuración de Mallas/Planes y Actualización de Tabla Cursos
-- ==============================================================================

-- 1. Desactivar validación de Foreign Keys temporalmente
PRAGMA foreign_keys = OFF;

-- 2. Limpiar tablas dependientes que van a ser reestructuradas
DELETE FROM horarios_detalle;
DELETE FROM horarios;
DELETE FROM docentes_curso;
DELETE FROM examenes;
DELETE FROM curso_horarios;
DELETE FROM cursos; 

-- 3. Eliminar la tabla vieja de mallas
DROP TABLE IF EXISTS mallas;

-- 4. Crear la tabla 'planes'
CREATE TABLE IF NOT EXISTS planes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    carrera INTEGER NOT NULL REFERENCES carreras(id) ON DELETE CASCADE ON UPDATE CASCADE,
    codigo VARCHAR(20) NOT NULL,
    UNIQUE (codigo, carrera)
);

-- 5. Crear la nueva tabla 'mallas'
CREATE TABLE mallas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    carrera INTEGER NOT NULL REFERENCES carreras(id) ON DELETE CASCADE ON UPDATE CASCADE,
    plan INTEGER NOT NULL REFERENCES planes(id) ON DELETE CASCADE ON UPDATE CASCADE,
    asignatura INTEGER NOT NULL REFERENCES asignaturas(id) ON DELETE CASCADE ON UPDATE CASCADE,
    semestre INTEGER NOT NULL DEFAULT 0,
    nivel INTEGER NOT NULL DEFAULT 0,

    UNIQUE (carrera, asignatura, plan)
);
CREATE INDEX IF NOT EXISTS idx_mallas_carrera_plan_semestre ON mallas(carrera, plan, semestre);

-- 6. Recrear tabla 'cursos' para añadir la columna 'turno' y actualizar el UNIQUE CONSTRAINT
DROP TABLE IF EXISTS cursos;

CREATE TABLE cursos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    malla INTEGER NOT NULL REFERENCES mallas(id) ON DELETE CASCADE ON UPDATE CASCADE,
    periodo INTEGER NOT NULL REFERENCES periodos(id) ON DELETE CASCADE ON UPDATE CASCADE,
    nombre TEXT NOT NULL,
    seccion VARCHAR(6) NOT NULL,
    turno VARCHAR(10) NOT NULL,
    tipo INTEGER NOT NULL DEFAULT 0 CHECK (tipo IN (0, 1)),

    -- Comité
    comite_presidente TEXT,
    comite_miembro1 TEXT,
    comite_miembro2 TEXT,

    fechas_sabados TEXT,

    -- Restricción UNIQUE actualizada para incluir 'turno'
    UNIQUE (nombre, malla, seccion, periodo, turno)
);

CREATE INDEX IF NOT EXISTS idx_cursos_malla_periodo ON cursos(malla, periodo);

-- 7. Crear tablas de Énfasis
CREATE TABLE IF NOT EXISTS enfasis (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    carrera INTEGER NOT NULL REFERENCES carreras(id) ON DELETE CASCADE ON UPDATE CASCADE,
    codigo VARCHAR(10) NOT NULL,
    nombre TEXT NOT NULL,
    UNIQUE (codigo, carrera)
);

CREATE TABLE IF NOT EXISTS enfasis_materia (
    malla INTEGER NOT NULL REFERENCES mallas(id) ON DELETE CASCADE ON UPDATE CASCADE,
    enfasis INTEGER NOT NULL REFERENCES enfasis(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (malla, enfasis)
);

-- 8. Reactivar Foreign Keys y verificar la consistencia del esquema
PRAGMA foreign_keys = ON;
PRAGMA foreign_key_check;
