create table laboratorios (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    malla INTEGER NOT NULL REFERENCES mallas(id) ON DELETE CASCADE ON UPDATE CASCADE,
    periodo INTEGER NOT NULL REFERENCES periodos(id) ON DELETE CASCADE ON UPDATE CASCADE,
    seccion VARCHAR(16) NOT NULL,

    UNIQUE (malla, seccion, periodo)
);

CREATE TABLE laboratorio_horarios (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    laboratorio INTEGER NOT NULL REFERENCES laboratorios(id) ON DELETE CASCADE,

    dia INTEGER NOT NULL CHECK (dia BETWEEN 1 AND 6), -- 1=lunes ... 6=sabado
    desde TEXT NOT NULL,
    hasta TEXT NOT NULL,
    aula TEXT,

    -- evita duplicados tipo "dos lunes iguales"
    UNIQUE (laboratorio, dia, desde)
);
CREATE INDEX idx_laboratorio_horarios_laboratorio ON laboratorio_horarios(laboratorio);
CREATE INDEX idx_laboratorio_horarios_dia ON laboratorio_horarios(dia);
