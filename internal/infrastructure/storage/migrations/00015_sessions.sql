-- +goose Up

-- Сессии входа: переживают перезапуск приложения. Хранится SHA-256 токена из куки,
-- а не сам токен: по копии базы нельзя войти под чужой сессией.
CREATE TABLE sessions (
    token_hash TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at INTEGER NOT NULL, -- Unix-время, секунды
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX idx_sessions_user_id ON sessions (user_id);
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);

-- +goose Down

DROP TABLE sessions;
