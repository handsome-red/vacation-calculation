-- +goose Up

INSERT INTO districts (code, title, is_active) VALUES
    ('kazan',              'г. Казань',               1),
    ('nizhnekamsk',        'Нижнекамский ТО',         1),
    ('naberezhnye_chelny', 'Набережночелнинский ТО',  1),
    ('almetyevsk',         'Альметьевский ТО',        1),
    ('nurlat',             'Нурлатский ТО',           1),
    ('zelenodolsk',        'Зеленодольский ТО',       1),
    ('buinsk',             'Буинский ТО',             1),
    ('bugulma',            'Бугульминский ТО',        1),
    ('elabuga',            'Елабужский ТО',           1),
    ('chistopol',          'Чистопольский ТО',        1),
    ('vysokogorsky',       'Высокогорский ТО',        1),
    ('laishevsky',         'Лаишевский ТО',           1),
    ('sabinsky',           'Сабинский ТО',            1)
ON CONFLICT (code) DO NOTHING;

INSERT INTO departments (code, title, district_code, is_active) VALUES
    -- Республиканские отделы (district_code = '')
    ('epidemiological_supervision',        'Отдел эпидемиологического надзора',                                     '', 1),
    ('sanitary_territory_protection',      'Отдел санитарной охраны территории',                                    '', 1),
    ('children_teens_hygiene_supervision', 'Отдел надзора по гигиене детей и подростков',                           '', 1),
    ('occupational_radiation_transport',   'Отдел надзора по гигиене труда, радиационной гигиене и на транспорте',  '', 1),
    ('food_hygiene_supervision',           'Отдел надзора по гигиене питания',                                      '', 1),
    ('medical_organizations_supervision',  'Отдел надзора за лечебно-профилактическими учреждениями',               '', 1),
    ('communal_hygiene_supervision',       'Отдел надзора по коммунальной гигиене',                                 '', 1),
    ('state_registration_licensing',       'Отдел государственной регистрации и лицензирования',                    '', 1),

    -- Территориальные отделы (с district_code)
    ('naberezhnye_chelny_to', 'Территориальный отдел в городе Набережные Челны, Актанышском районе',                 'naberezhnye_chelny', 1),
    ('almetyevsk_to',         'Территориальный отдел в Альметьевском, Заинском, Лениногорском районах',              'almetyevsk',         1),
    ('nurlat_to',             'Территориальный отдел в Нурлатском, Аксубаевском районах',                            'nurlat',             1),
    ('nizhnekamsk_to',        'Территориальный отдел в Нижнекамском районе и г. Нижнекамск',                         'nizhnekamsk',        1),
    ('zelenodolsk_to',        'Территориальный отдел в Зеленодольском, Верхнеуслонском, Камско-Устьинском районах',  'zelenodolsk',        1),
    ('buinsk_to',             'Территориальный отдел в Буинском, Тетюшском районах',                                 'buinsk',             1),
    ('bugulma_to',            'Территориальный отдел в Бугульминском, Азнакаевском, Бавлинском районах',             'bugulma',            1),
    ('elabuga_to',            'Территориальный отдел в Елабужском, Агрызском, Менделеевском, Мамадышском районах',  'elabuga',            1),
    ('chistopol_to',          'Территориальный отдел в Чистопольском, Спасском районах',                             'chistopol',          1),
    ('arsky_to',              'Территориальный отдел в Арском, Высокогорском районах',                               'vysokogorsky',       1),
    ('laishevsky_to',         'Территориальный отдел в Лаишевском, Рыбно-Слободском районах',                        'laishevsky',         1),
    ('sabinsky_to',           'Территориальный отдел в Сабинском, Кукморском районах',                               'sabinsky',           1)
ON CONFLICT (code) DO NOTHING;

-- +goose Down

DELETE FROM departments WHERE code IN (
    'epidemiological_supervision',
    'sanitary_territory_protection',
    'children_teens_hygiene_supervision',
    'occupational_radiation_transport',
    'food_hygiene_supervision',
    'medical_organizations_supervision',
    'communal_hygiene_supervision',
    'state_registration_licensing',
    'naberezhnye_chelny_to', 'almetyevsk_to', 'nurlat_to', 'nizhnekamsk_to',
    'zelenodolsk_to', 'buinsk_to', 'bugulma_to', 'elabuga_to',
    'chistopol_to', 'arsky_to', 'laishevsky_to', 'sabinsky_to'
);

DELETE FROM districts WHERE code IN (
    'kazan', 'nizhnekamsk', 'naberezhnye_chelny', 'almetyevsk', 'nurlat',
    'zelenodolsk', 'buinsk', 'bugulma', 'elabuga', 'chistopol',
    'vysokogorsky', 'laishevsky', 'sabinsky'
);