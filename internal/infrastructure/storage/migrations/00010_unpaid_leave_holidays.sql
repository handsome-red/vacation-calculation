-- +goose Up

-- Административный отпуск (без сохранения заработной платы): новый вид 'unpaid'.
-- CHECK в SQLite не изменить без пересоздания таблицы. Ссылки на vacations проверяются
-- в конце транзакции, а история переносов сохраняется: DROP TABLE каскадно удалил бы её.
PRAGMA defer_foreign_keys = ON;

CREATE TABLE _vacation_transfers_backup AS SELECT * FROM vacation_transfers;

CREATE TABLE vacations_new (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_date       TEXT NOT NULL,
    end_date         TEXT NOT NULL,
    kind             TEXT NOT NULL CHECK (kind IN ('base', 'irregular', 'seniority', 'unpaid')),
    work_year_start  TEXT NOT NULL,
    status           TEXT NOT NULL,
    created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    -- ссылка на саму новую таблицу: при RENAME SQLite заменит имя на vacations
    transferred_from TEXT REFERENCES vacations_new(id)
);

INSERT INTO vacations_new (id, user_id, start_date, end_date, kind, work_year_start, status, created_at, updated_at, transferred_from)
SELECT id, user_id, start_date, end_date, kind, work_year_start, status, created_at, updated_at, transferred_from
FROM vacations;

DROP TABLE vacations;
ALTER TABLE vacations_new RENAME TO vacations;

CREATE INDEX idx_vacations_user_id   ON vacations(user_id);
CREATE INDEX idx_vacations_work_year ON vacations(user_id, work_year_start);

INSERT OR IGNORE INTO vacation_transfers SELECT * FROM _vacation_transfers_backup;
DROP TABLE _vacation_transfers_backup;

-- Праздники Республики Татарстан и постоянные праздники 2027 года (ст. 112 ТК РФ).
-- Переносы выходных и Ураза-байрам/Курбан-байрам администратор добавляет сам.
INSERT OR IGNORE INTO work_calendar_days (date, day_type, name) VALUES
    ('2026-08-30', 'holiday', 'День Республики Татарстан'),
    ('2027-01-01', 'holiday', 'Новогодние каникулы'),
    ('2027-01-02', 'holiday', 'Новогодние каникулы'),
    ('2027-01-03', 'holiday', 'Новогодние каникулы'),
    ('2027-01-04', 'holiday', 'Новогодние каникулы'),
    ('2027-01-05', 'holiday', 'Новогодние каникулы'),
    ('2027-01-06', 'holiday', 'Новогодние каникулы'),
    ('2027-01-07', 'holiday', 'Рождество Христово'),
    ('2027-01-08', 'holiday', 'Новогодние каникулы'),
    ('2027-02-23', 'holiday', 'День защитника Отечества'),
    ('2027-03-08', 'holiday', 'Международный женский день'),
    ('2027-05-01', 'holiday', 'Праздник Весны и Труда'),
    ('2027-05-09', 'holiday', 'День Победы'),
    ('2027-06-12', 'holiday', 'День России'),
    ('2027-08-30', 'holiday', 'День Республики Татарстан'),
    ('2027-11-04', 'holiday', 'День народного единства');

-- +goose Down

DELETE FROM work_calendar_days WHERE date = '2026-08-30' OR date BETWEEN '2027-01-01' AND '2027-12-31';
