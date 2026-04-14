-- Relaciones dependientes primero

DROP TABLE IF EXISTS docentes_curso;
DROP TABLE IF EXISTS examenes;
DROP TABLE IF EXISTS curso_horarios;

-- Cursos depende de mallas y periodos
DROP TABLE IF EXISTS cursos;

-- Mallas depende de carreras y asignaturas
DROP TABLE IF EXISTS mallas;

-- Tablas base de catálogo
DROP TABLE IF EXISTS asignaturas;

DROP TABLE IF EXISTS periodos;

DROP TABLE IF EXISTS docentes;

DROP TABLE IF EXISTS departamentos;

DROP TABLE IF EXISTS carreras;
