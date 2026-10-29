-- +goose Up

-- Уведомление о начале отпуска получают сотрудник и отдел кадров; отправка
-- учитывается по каждому получателю, чтобы неудачное письмо повторялось только ему.
-- Таблица пересоздаётся: до этой миграции в ней не было ни одной записи с получателем.
DROP TABLE vacation_reminders;

CREATE TABLE vacation_reminders (
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_date  TEXT NOT NULL,
    end_date    TEXT NOT NULL,
    recipient   TEXT NOT NULL,  -- 'employee' или 'hr:<адрес>'
    email       TEXT NOT NULL,
    sent_at     TEXT NOT NULL,
    PRIMARY KEY (user_id, start_date, recipient)
);

-- +goose Down

DROP TABLE vacation_reminders;

CREATE TABLE vacation_reminders (
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_date  TEXT NOT NULL,
    end_date    TEXT NOT NULL,
    email       TEXT NOT NULL,
    sent_at     TEXT NOT NULL,
    PRIMARY KEY (user_id, start_date)
);
