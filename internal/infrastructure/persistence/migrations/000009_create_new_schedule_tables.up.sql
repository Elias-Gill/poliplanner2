-- +migrate Up
-- Tabla principal de horarios del usuario
CREATE TABLE horarios (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    usuario_id INTEGER NOT NULL,
    titulo TEXT NOT NULL DEFAULT 'Mi horario',
    creado_en DATETIME DEFAULT (datetime('now')),

    FOREIGN KEY (usuario_id) REFERENCES users(user_id) ON DELETE CASCADE,
    UNIQUE (usuario_id, titulo)
);

-- Detalle de cursos en cada horario
CREATE TABLE horarios_detalle (
    horario_id INTEGER NOT NULL,
    curso_id INTEGER NOT NULL,
    PRIMARY KEY (horario_id, curso_id),

    FOREIGN KEY (horario_id) REFERENCES horarios(id) ON DELETE CASCADE,
    FOREIGN KEY (curso_id) REFERENCES cursos(id) ON DELETE CASCADE
);
