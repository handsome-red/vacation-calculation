-- +goose Up

-- Сохранённые рабочие годы: завершившиеся (их границы больше не пересчитываются) и
-- с границами, заданными администратором. Годы идут подряд с первого; следующие
-- считаются от конца последнего сохранённого. Заменяет work_year_ends: завершившиеся
-- годы сохранит приложение при запуске.
CREATE TABLE work_years (
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    number     INTEGER NOT NULL CHECK (number >= 1),
    start_date TEXT NOT NULL,
    end_date   TEXT NOT NULL CHECK (end_date >= start_date), -- последний день, включительно
    manual     INTEGER NOT NULL DEFAULT 0 CHECK (manual IN (0, 1)),
    PRIMARY KEY (user_id, number)
);

DROP TABLE work_year_ends;

-- +goose Down

CREATE TABLE work_year_ends (
    user_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    number   INTEGER NOT NULL CHECK (number >= 1),
    last_day TEXT NOT NULL,
    PRIMARY KEY (user_id, number)
);

INSERT INTO work_year_ends (user_id, number, last_day)
SELECT user_id, number, end_date FROM work_years WHERE manual = 1;

DROP TABLE work_years;
