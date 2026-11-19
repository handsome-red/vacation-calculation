-- +goose Up

-- Дни, использованные до начала учёта в системе: даты таких отпусков неизвестны,
-- известен только остаток по рабочему году.
CREATE TABLE vacation_opening_usage (
    user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    work_year_start TEXT NOT NULL,
    base            INTEGER NOT NULL DEFAULT 0 CHECK (base >= 0),
    irregular       INTEGER NOT NULL DEFAULT 0 CHECK (irregular >= 0),
    seniority       INTEGER NOT NULL DEFAULT 0 CHECK (seniority >= 0),
    PRIMARY KEY (user_id, work_year_start)
);

-- +goose Down

DROP TABLE vacation_opening_usage;
