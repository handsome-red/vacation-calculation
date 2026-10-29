-- +goose Up

-- Стаж (в днях) до приёма на работу
ALTER TABLE users ADD COLUMN experience INTEGER NOT NULL DEFAULT 0;

-- Исключения производственного календаря (праздники, рабочие выходные)
CREATE TABLE work_calendar_days (
    date      TEXT PRIMARY KEY,
    day_type  TEXT NOT NULL CHECK (day_type IN ('holiday', 'working', 'weekend')),
    name      TEXT NOT NULL DEFAULT ''
);

-- Заменена таблицей work_calendar_days
DROP TABLE holidays;

-- +goose Down

CREATE TABLE holidays (
    date  TEXT PRIMARY KEY,
    name  TEXT NOT NULL
);

DROP TABLE work_calendar_days;

ALTER TABLE users DROP COLUMN experience;
