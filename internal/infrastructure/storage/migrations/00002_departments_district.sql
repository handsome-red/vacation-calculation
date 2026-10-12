-- +goose Up
ALTER TABLE departments ADD COLUMN district_code TEXT NOT NULL DEFAULT '';

-- +goose Down
-- SQLite: drop column требует версии >= 3.35 и перестройки таблицы.
-- Проще оставить без down или пересоздать таблицу.