-- +goose Up

-- Журнал изменений. Без внешних ключей: записи остаются после удаления сотрудника,
-- имена сохранены на момент действия.
CREATE TABLE audit_log (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    at            TEXT NOT NULL, -- RFC 3339, UTC
    actor_id      TEXT NOT NULL,
    actor_name    TEXT NOT NULL,
    target_id     TEXT NOT NULL DEFAULT '',
    target_name   TEXT NOT NULL DEFAULT '',
    district_code TEXT NOT NULL DEFAULT '',
    action        TEXT NOT NULL,
    details       TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_audit_log_target ON audit_log (target_id, id);
CREATE INDEX idx_audit_log_district ON audit_log (district_code, id);

-- +goose Down

DROP TABLE audit_log;
