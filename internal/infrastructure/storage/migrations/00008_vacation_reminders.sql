-- +goose Up

-- Журнал уведомлений о начале отпуска: по одному на сотрудника и дату начала,
-- чтобы письмо не уходило повторно. Перенос на другую дату начала — новое уведомление.
CREATE TABLE vacation_reminders (
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_date  TEXT NOT NULL,
    end_date    TEXT NOT NULL,
    email       TEXT NOT NULL,
    sent_at     TEXT NOT NULL,
    PRIMARY KEY (user_id, start_date)
);

-- +goose Down

DROP TABLE vacation_reminders;
