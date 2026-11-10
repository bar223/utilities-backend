# ER-диаграмма — данные для StarUML

Готовая диаграмма — `docs/diagrams/utilities.mdj` (ветка `lab3-rest-api`,
вкладка «ER Diagram»). Оформление: без сетки (View → Grid off), под лист
А4, таблицы компактными прямоугольниками.

Обозначения в колонке «Ключ»: PK — первичный ключ, FK — внешний ключ,
U — уникальный, N — необязательный (NULL). Без N — NOT NULL.

## Таблица 1: utilities

| Ключ | Поле | Тип |
|---|---|---|
| PK | id | BIGINT |
| | name | VARCHAR(100) |
| N | description | VARCHAR(500) |
| | status | VARCHAR(15) |
| | image_url | VARCHAR(255) |
| | video_url | VARCHAR(255) |
| N | tariff_rate | NUMERIC(10,2) |
| N | consumption_norm | NUMERIC(10,3) |
| | measurement_unit | VARCHAR(20) |
| | created_at | TIMESTAMPTZ |
| N | published_at | TIMESTAMPTZ |
| FK | creator_id | BIGINT |

Необязательных полей ровно 4 — те, что заполняются только при публикации:
`description`, `tariff_rate`, `consumption_norm`, `published_at`.
`image_url`, `video_url` обязательные (по «Далее» записываются неверные
ссылки), `measurement_unit` обязательный с `DEFAULT ''` — формат создания
черновика (только название) не меняется.

## Таблица 2: likes

| Ключ | Поле | Тип |
|---|---|---|
| PK | id | BIGINT |
| FK | user_id | BIGINT |
| FK | utility_id | BIGINT |

Уникальная пара (user_id, utility_id) — один пользователь не может
лайкнуть одну услугу дважды.

## Таблица 3: users

| Ключ | Поле | Тип |
|---|---|---|
| PK | id | BIGINT |
| U | login | VARCHAR(50) |
| | password | VARCHAR(100) |
| | is_moderator | BOOLEAN |

## Связи

- `users` 1 — 0..* `utilities` по `creator_id`
- `users` 1 — 0..* `likes` по `user_id`
- `utilities` 1 — 0..* `likes` по `utility_id`

Каскадное удаление не используется — удаление только логическое, через
поле `status`.

TIMESTAMPTZ — TIMESTAMP WITH TIME ZONE, как в PostgreSQL после GORM AutoMigrate.
Отметка U у отдельных user_id и utility_id не ставится: уникален составной индекс idx_user_utility (user_id, utility_id).
Связи неидентифицирующие: у likes собственный PK id; внешние ключи не входят в PK.
