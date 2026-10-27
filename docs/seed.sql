-- Тестовые данные для защиты: пользователи, услуги во всех трёх статусах,
-- лайки. Запуск после `go run ./cmd/migrate`:
--   Get-Content docs\seed.sql | docker exec -i utilities-postgres psql -U postgres -d utilities
-- Повторный запуск очищает таблицы и заливает данные заново.
--
-- Пароли — bcrypt-хэш (pgcrypto), поэтому вход в ЛР-4 работает:
-- creator / 123 (создатель всех услуг, id = 1), neighbor / 123.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

TRUNCATE likes, utilities, users RESTART IDENTITY;

INSERT INTO users (login, password, is_moderator) VALUES
    ('creator',  crypt('123', gen_salt('bf', 10)), false),
    ('neighbor', crypt('123', gen_salt('bf', 10)), false);

INSERT INTO utilities (name, description, status, image_url, video_url, tariff_rate, consumption_norm,
                       measurement_unit, created_at, published_at, creator_id) VALUES
    ('Электроэнергия',
     'Оплата электрической энергии, потребляемой квартирой: освещение, розетки и бытовые приборы.',
     'published', 'http://localhost:9000/utilities/electricity.jpg', 'http://localhost:9000/utilities/electricity.mp4',
     6.10, 95.5, 'кВт·ч', now() - interval '10 days', now() - interval '10 days', 1),
    ('Отопление',
     'Тепловая энергия на отопление жилого помещения в отопительный период.',
     'published', 'http://localhost:9000/utilities/heating.jpg', 'http://localhost:9000/utilities/heating.mp4',
     2450.00, 0.016, 'Гкал', now() - interval '9 days', now() - interval '9 days', 1),
    ('Холодное водоснабжение',
     'Подача холодной питьевой воды в квартиру по показаниям счётчика.',
     'published', 'http://localhost:9000/utilities/cold-water.jpg', 'http://localhost:9000/utilities/cold-water.mp4',
     47.24, 4.9, 'м³', now() - interval '8 days', now() - interval '8 days', 1),
    ('Газоснабжение',
     'Оплата природного газа, который используется для плиты и проточного водонагревателя.',
     'published', 'http://localhost:9000/utilities/gas-supply.jpg', 'http://localhost:9000/utilities/gas-supply.mp4',
     8.20, 10.0, 'м³', now() - interval '7 days', now() - interval '7 days', 1),
    ('Вывоз ТКО',
     'Вывоз и утилизация твёрдых коммунальных отходов, начисляется на зарегистрированного жильца.',
     'published', 'http://localhost:9000/utilities/waste-removal.jpg', 'http://localhost:9000/utilities/waste-removal.mp4',
     12.40, 0.18, 'м³', now() - interval '6 days', now() - interval '6 days', 1),
    ('Капитальный ремонт',
     'Взнос на капитальный ремонт общего имущества многоквартирного дома.',
     'published', 'http://localhost:9000/utilities/capital-repair.jpg', 'http://localhost:9000/utilities/capital-repair.mp4',
     24.20, 1.0, 'м²', now() - interval '5 days', now() - interval '5 days', 1),
    ('Горячее водоснабжение',
     'Удалённая услуга — в интерфейсе и API не показывается.',
     'deleted', 'http://localhost:9000/utilities/hot-water.jpg', 'http://localhost:9000/utilities/hot-water.mp4',
     250.00, 3.6, 'м³', now() - interval '4 days', now() - interval '4 days', 1),
    ('Домофон',
     '', 'draft', 'http://localhost:9000/utilities/utility-8-image.jpg', 'http://localhost:9000/utilities/utility-8-video.mp4',
     NULL, NULL, '', now() - interval '1 day', NULL, 2);

INSERT INTO likes (user_id, utility_id) VALUES
    (1, 1), (2, 1), (1, 3), (2, 4), (1, 6), (2, 6), (2, 2);
