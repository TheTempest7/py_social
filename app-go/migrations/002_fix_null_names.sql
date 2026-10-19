-- SQLite не поддерживает ALTER COLUMN — придётся пересоздавать таблицу
CREATE TABLE users_new (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT NOT NULL UNIQUE,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    first_name    TEXT NOT NULL DEFAULT '',   -- 👈 fix
    last_name     TEXT NOT NULL DEFAULT '',   -- 👈 fix
    is_active     BOOLEAN DEFAULT 1,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users_new 
    SELECT id, email, username, password_hash, 
           COALESCE(first_name, ''), COALESCE(last_name, ''),
           is_active, created_at, updated_at
    FROM users;

DROP TABLE users;
ALTER TABLE users_new RENAME TO users;