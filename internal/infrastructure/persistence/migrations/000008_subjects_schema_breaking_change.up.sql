-- Carreras
CREATE TABLE carreras (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- acronimo de la carrera (ej: IIN, ISP). Debe de ir en mayusculas
    siglas VARCHAR(6) NOT NULL,
    UNIQUE (siglas)
);

-- Docentes
CREATE TABLE docentes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- nombre crudo y sin normalizacion
    nombre TEXT NOT NULL,
    apellido TEXT NOT NULL,
    -- se utiliza como principal llave de busqueda para la base de datos de docentes
    correo TEXT NOT NULL,
    -- key que se utiliza para ciertas busquedas rapidas. De uso secundario.
    -- Es la combinacion del primer nombre y el primer apellido. Ej: juan_gonzales
    search_key TEXT NOT NULL,
    -- Los docentes se diferencian ultimamente por su correo electronico, asi por mas
    -- de que
    -- existan nombres iguales, seran unicos segun su correo.
    UNIQUE (correo)
);
CREATE INDEX idx_docentes_search_key ON docentes(search_key);

-- Departamentos
CREATE TABLE departamentos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    siglas VARCHAR(6) UNIQUE NOT NULL,
    UNIQUE (siglas)
);

-- Asignaturas (concepto académico, estable. Misma asignatura puede pertenecer a
-- varias mallas)
CREATE TABLE asignaturas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre TEXT NOT NULL,
    departamento INTEGER NOT NULL REFERENCES departamentos(id) ON DELETE CASCADE ON UPDATE CASCADE,
    -- No deberian de haber asignaturas con el mismo nombre, puesto que son unidades
    -- unicas y estables en el tiempo
    UNIQUE (nombre)
);
CREATE INDEX idx_asignaturas_departamento ON asignaturas(departamento);

-- Malla curricular (asignatura por carrera). Semestre puede ser rellenado con
-- metadatos luego
CREATE TABLE mallas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    carrera INTEGER NOT NULL REFERENCES carreras(id) ON DELETE CASCADE ON UPDATE CASCADE,
    asignatura INTEGER NOT NULL REFERENCES asignaturas(id) ON DELETE CASCADE ON UPDATE CASCADE,
    semestre INTEGER NOT NULL DEFAULT 0,
    nivel INTEGER NOT NULL DEFAULT 0,

    -- Asegura que no se repitan asignaturas en la misma carrera
    UNIQUE (carrera, asignatura)
);
-- busqueda por carrera
CREATE INDEX idx_mallas_carrera_semestre ON mallas(carrera);

-- Períodos académicos
CREATE TABLE periodos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL,
    periodo INTEGER NOT NULL CHECK (periodo IN (1, 2)),
    UNIQUE (year, periodo)
);

-- Cursos (asignatura dictada en un período. Diferenciados por secciones)
CREATE TABLE cursos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    malla INTEGER NOT NULL REFERENCES mallas(id) ON DELETE CASCADE ON UPDATE CASCADE,
    periodo INTEGER NOT NULL REFERENCES periodos(id) ON DELETE CASCADE ON UPDATE CASCADE,
    nombre TEXT NOT NULL,
    seccion VARCHAR(6) NOT NULL,
    tipo INTEGER NOT NULL DEFAULT 0 CHECK (tipo IN (0, 1)),

    -- Comité (está bien dejarlo acá)
    comite_presidente TEXT,
    comite_miembro1 TEXT,
    comite_miembro2 TEXT,

    UNIQUE (malla, seccion, periodo)
);
CREATE INDEX idx_cursos_malla_periodo ON cursos(malla, periodo);

CREATE TABLE curso_horarios (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    curso_id INTEGER NOT NULL REFERENCES cursos(id) ON DELETE CASCADE,

    dia INTEGER NOT NULL CHECK (dia BETWEEN 1 AND 6), -- 1=lunes ... 6=sabado
    desde TEXT NOT NULL,
    hasta TEXT NOT NULL,
    aula TEXT,

    -- evita duplicados tipo "dos lunes iguales"
    UNIQUE (curso_id, dia, desde)
);
CREATE INDEX idx_curso_horarios_curso ON curso_horarios(curso_id);
CREATE INDEX idx_curso_horarios_dia ON curso_horarios(dia);

CREATE TABLE examenes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    curso_id INTEGER NOT NULL REFERENCES cursos(id) ON DELETE CASCADE,

    tipo TEXT NOT NULL, -- 'partial', 'final'
    instancia INTEGER NOT NULL, -- 1, 2, etc

    fecha DATE NOT NULL,
    hora TEXT NOT NULL,
    aula TEXT,

    revision_fecha DATE,
    revision_hora TEXT,

    UNIQUE (curso_id, tipo, instancia)
);
CREATE INDEX idx_examenes_curso ON examenes(curso_id);

-- Tabla con los docentes del curso, porque hay cursos con mas de un docente
CREATE TABLE docentes_curso(
    id_docente INTEGER NOT NULL REFERENCES docentes(id) ON DELETE CASCADE ON UPDATE CASCADE,
    id_curso INTEGER NOT NULL REFERENCES cursos(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (id_docente, id_curso)
);
