CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    recovery_token_hash TEXT,
    recovery_token_expiration DATETIME,
    recovery_token_used BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_users_name ON users(username);
