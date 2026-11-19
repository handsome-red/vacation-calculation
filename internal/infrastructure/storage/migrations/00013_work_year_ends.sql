-- +goose Up

-- Конец рабочего года, заданный администратором. Год определяется порядковым номером:
-- следующий год начинается на другой день, а номер остаётся прежним.
CREATE TABLE work_year_ends (
    user_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    number   INTEGER NOT NULL CHECK (number >= 1),
    last_day TEXT NOT NULL,
    PRIMARY KEY (user_id, number)
);

-- +goose Down

DROP TABLE work_year_ends;
