-- +goose Up

-- Отпуск хранится кусками: каждый кусок списан из квоты одного вида
-- одного рабочего года. Таблица пересоздаётся: до этой миграции в ней были
-- только тестовые данные без рабочего года-источника.
DROP INDEX IF EXISTS idx_vacations_user_id;
DROP TABLE vacations;

CREATE TABLE vacations (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_date       TEXT NOT NULL,
    end_date         TEXT NOT NULL,
    kind             TEXT NOT NULL CHECK (kind IN ('base', 'irregular', 'seniority')),
    work_year_start  TEXT NOT NULL,  -- начало рабочего года, из которого взяты дни
    status           TEXT NOT NULL,
    created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

CREATE INDEX idx_vacations_user_id   ON vacations(user_id);
CREATE INDEX idx_vacations_work_year ON vacations(user_id, work_year_start);

-- +goose Down

DROP INDEX idx_vacations_work_year;
DROP INDEX idx_vacations_user_id;
DROP TABLE vacations;

CREATE TABLE vacations (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_date  TEXT NOT NULL,
    end_date    TEXT NOT NULL,
    status      TEXT NOT NULL,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

CREATE INDEX idx_vacations_user_id ON vacations(user_id);
