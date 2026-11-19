-- +goose Up

-- Отдел был только в локальной базе: без него сотрудники ОИТ не находятся (JOIN departments).
INSERT INTO departments (code, title, district_code, is_active) VALUES
    ('information_technology', 'Отдел информационных технологий', 'kazan', 1)
ON CONFLICT (code) DO NOTHING;

-- +goose Down
-- Запись могла существовать до миграции и использоваться сотрудниками: не удаляем.
