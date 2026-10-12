-- +goose Up

-- Справочники
CREATE TABLE departments (
    code       TEXT PRIMARY KEY,
    title      TEXT NOT NULL,
    is_active  INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE districts (
    code       TEXT PRIMARY KEY,
    title      TEXT NOT NULL,
    is_active  INTEGER NOT NULL DEFAULT 1
);

-- Аутентификация
CREATE TABLE auth_users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL,
    status        TEXT NOT NULL,
    created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

CREATE INDEX idx_auth_users_email ON auth_users(LOWER(email));

-- Профиль сотрудника
CREATE TABLE users (
    id                TEXT PRIMARY KEY,
    status            TEXT NOT NULL,
    last_name         TEXT NOT NULL,
    first_name        TEXT NOT NULL,
    middle_name       TEXT NOT NULL DEFAULT '',
    birth_date        TEXT NOT NULL,
    position          TEXT NOT NULL,
    hired_at          TEXT NOT NULL,
    department_code   TEXT NOT NULL,
    district_code     TEXT NOT NULL,
    workday_duration  INTEGER NOT NULL DEFAULT 8,
    email             TEXT NOT NULL UNIQUE,
    is_invalid        INTEGER NOT NULL DEFAULT 0,
    password          TEXT NOT NULL,
    superadmin        INTEGER NOT NULL DEFAULT 0 CHECK (superadmin IN (0, 1)),
    created_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

CREATE INDEX idx_users_email           ON users(LOWER(email));
CREATE INDEX idx_users_department_code ON users(department_code);
CREATE INDEX idx_users_district_code   ON users(district_code);

-- Отпуска
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

-- Смены
CREATE TABLE shifts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind       TEXT NOT NULL,
    date_from  TEXT NOT NULL,
    date_to    TEXT NOT NULL
);

CREATE INDEX idx_shifts_user_id ON shifts(user_id);

-- Праздники
CREATE TABLE holidays (
    date  TEXT PRIMARY KEY,
    name  TEXT NOT NULL
);

-- +goose Down

DROP TABLE holidays;
DROP INDEX idx_shifts_user_id;
DROP TABLE shifts;
DROP INDEX idx_vacations_user_id;
DROP TABLE vacations;
DROP INDEX idx_users_district_code;
DROP INDEX idx_users_department_code;
DROP INDEX idx_users_email;
DROP TABLE users;
DROP INDEX idx_auth_users_email;
DROP TABLE auth_users;
DROP TABLE districts;
DROP TABLE departments;