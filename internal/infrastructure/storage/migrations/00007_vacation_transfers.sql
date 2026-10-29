-- +goose Up

-- Запись отпуска, созданная переносом, ссылается на исходный (перенесённый) отпуск.
ALTER TABLE vacations ADD COLUMN transferred_from TEXT REFERENCES vacations(id);

-- История переносов: какой отпуск, с каких дат, на какие, почему и кто перенёс.
CREATE TABLE vacation_transfers (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vacation_id  TEXT NOT NULL REFERENCES vacations(id) ON DELETE CASCADE,
    old_start    TEXT NOT NULL,
    old_end      TEXT NOT NULL,
    new_start    TEXT NOT NULL,
    new_end      TEXT NOT NULL,
    reason       TEXT NOT NULL,
    created_by   TEXT NOT NULL REFERENCES users(id),
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

CREATE INDEX idx_vacation_transfers_user_id ON vacation_transfers(user_id);

-- +goose Down

DROP INDEX idx_vacation_transfers_user_id;
DROP TABLE vacation_transfers;
ALTER TABLE vacations DROP COLUMN transferred_from;
