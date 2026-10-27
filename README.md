# Utilities — backend

Учебный проект по дисциплине **«Разработка Интернет Приложений» (РИП), 2026**.

Тема варианта — **«Квитанция по квартплате»**. Предметная область: начисления за
коммунальные ресурсы, из которых складывается месячная квитанция жильца.

Фронтенд — репозиторий [utility-bills-frontend](https://github.com/bar223/utility-bills-frontend).

## Лабораторная работа 2 — база данных

Ветка: `lab2-database`

Тема лабораторной — структура БД и её подключение к бэкенду. Коллекция
в памяти из лабы 1 заменена на PostgreSQL через GORM.

Сделано:

- 3 таблицы (каскадное удаление не используется, статусы — soft delete):
  `utilities`, `likes` (м-м пользователь-услуга), `users`;
- поле по теме `consumptionNorm` (норматив потребления) добавлено к уже
  существующему `tariffRate` — оба поля числовые, `tariffRate`
  используется для фильтра на плитке;
- 6 HTTP-методов: 3 GET через ORM (плитка `/`, лента
  `/utilities/feed/:id`, черновик `/utilities/draft`), создание черновика
  через ORM (`POST /utilities/draft`), публикация через ORM
  (`POST /utilities/:id/publish`), логическое удаление прямым SQL `UPDATE`
  без ORM (`POST /utilities/:id/delete`);
- на странице «Добавление» сверху друг под другом видео и фото — на обоих
  шагах («Далее» и «Опубликовать»); выбор файлов из проводника — только до
  «Далее». Файлы на сервер не передаются, в обязательные поля `image_url`
  и `video_url` записываются неверные ссылки — вместо них показываются
  видео и картинка по умолчанию, которые хранятся на SSR-сервере
  (`resources/`), а не в MinIO. Подмена делается без JavaScript:
  `<object>` с вложенной картинкой и второй `<source>` у `<video>`;
- миграции — `cmd/migrate` (`go run ./cmd/migrate`);
- в БД гарантированно есть услуги во всех трёх статусах: `draft`,
  `published` (×6), `deleted`.

Инфраструктура — PostgreSQL и Adminer подняты в `docker-compose.yml`
рядом с MinIO:

```bash
docker compose up -d
go run ./cmd/migrate      # создаёт схему
go run ./cmd/utilities
```

- PostgreSQL — `localhost:5433` (порт сдвинут с 5432 — на машине уже стоит
  нативный PostgreSQL, чтобы не занимать его порт)
- Adminer — http://localhost:8081 (система: PostgreSQL, сервер: `postgres`,
  логин/пароль — из `.env`, база — `utilities`)

### ER-диаграмма

Таблицы для диаграммы в StarUML:

- **utilities** — `id` PK, `name`, `description` N, `status`,
  `image_url`, `video_url`, `tariff_rate` N, `consumption_norm` N,
  `measurement_unit`, `created_at`, `published_at` N, `creator_id` FK → users.id
  (N — необязательное поле, всего 4; подробно — `docs/lab2/er-diagram-spec.md`)
- **likes** — `id` PK, `user_id` FK → users.id, `utility_id`
  FK → utilities.id (уникальная пара)
- **users** — `id` PK, `login`, `password`, `is_moderator`

Связи: `users` 1 — * `utilities` (создатель), `users` 1 — *
`likes` * — 1 `utilities`.

## Лабораторная работа 1 — SSR

Ветка: `lab1-ssr`

Тема лабораторной — серверный рендеринг (SSR): три страницы приложения
(«Плитка», «Лента», «Добавление») собираются на сервере через
`html/template` и отдаются готовым HTML, без клиентского JavaScript.

Сделано:

- веб-сервис на Go (Gin + `html/template`), архитектура по методическим
  указаниям курса: `cmd` → `internal/api` → `internal/app/handler` →
  `internal/app/repository`;
- доменная коллекция `CommunalResource` (с лабы 2 — `Utility`) в памяти, без БД (требование
  лабы 1) — 6 опубликованных услуг + 1 черновик;
- три GET-метода: список с фильтром по тарифу, лента по ID с переходом
  «следующая услуга», получение черновика;
- вёрстка по макету Figma, изображения и видео — из MinIO;
- поднята инфраструктура объектного хранилища MinIO через
  `docker-compose.yml`.

## Доменная модель

### CommunalResource / Utility — коммунальный ресурс (услуга)

Основная сущность варианта. Одна запись — один вид коммунального ресурса,
по которому производится начисление. В коде лабы 1 называлась
`CommunalResource`, с лабы 2 переименована в `Utility` (таблица БД —
`utilities`).

| Поле | Тип | Описание |
|---|---|---|
| `name` | string | Наименование ресурса |
| `tariffRate` | number | Тариф — стоимость за единицу измерения |
| `measurementUnit` | string | Единица измерения |
| `description` | text | Подробное описание ресурса |
| `imageUrl` | string | Ссылка на изображение (хранится в MinIO) |
| `videoUrl` | string | Ссылка на видеоматериал (хранится в MinIO) |
| `likes` | ID[] | Массив идентификаторов пользователей, отметивших ресурс |
| `status` | enum | `draft` / `published` / `deleted` |

Поле `status` задаёт жизненный цикл записи: `draft` — черновик, не виден
пользователям; `published` — опубликован; `deleted` — логическое удаление,
запись из базы не вычищается.

Наполнение справочника по варианту: Электроэнергия, Отопление, Холодное
водоснабжение, Газоснабжение, Вывоз ТКО, Капитальный ремонт.

### MonthlyReceipt — месячная квитанция

Квитанция за расчётный период по одному лицевому счёту. Агрегирует записи
о потреблении и содержит итоговую сумму к оплате. Появится в следующих
лабораторных работах вместе с БД.

### ConsumptionRecord — запись о потреблении

Связующая сущность между `MonthlyReceipt` и `CommunalResource`: объём
потребления конкретного ресурса в конкретной квитанции и рассчитанная
по тарифу сумма.

Связь: `MonthlyReceipt` 1 — * `ConsumptionRecord` * — 1 `CommunalResource`.

## Структура проекта

```
cmd/utilities/           точка входа (main.go)
cmd/migrate/             миграции (AutoMigrate по моделям ds)
internal/
  pkg/                   сборка приложения: роутер, запуск сервера
  app/
    config/              конфигурация (config/config.toml)
    dsn/                 строка подключения к PostgreSQL из .env
    ds/                  модели: Utility, User, Like
    handler/             обработчики HTTP-запросов
    repository/          доступ к БД (Utility, User, Like)
templates/               html/template-шаблоны трёх страниц
resources/styles/        CSS
resources/img, video/    фото и видео по умолчанию (не в MinIO)
docs/lab2/               скриншоты, сценарий защиты, ER-спецификация
```

## Запуск

```bash
cp .env.example .env
docker compose up -d   # поднимает MinIO
go run ./cmd/utilities
```

Приложение слушает `:8080`:

- `/` — плитка, список опубликованных услуг, фильтр по тарифу — слайдер
  с кнопкой «Применить» (`?tariff=`)
- `/utilities/feed/:id` — лента, переход к следующей услуге через
  `?next=true` (после последней — снова первая)
- `/utilities/draft` — черновик

## Инфраструктура: MinIO

Объектное хранилище MinIO разворачивается в Docker Desktop по методическим
указаниям курса ([Установка и администрирование Minio](https://github.com/iu5git/Web/blob/main/tutorials/minio/MinIO_bucket_setup_ReadMe.md)).

- контейнер — `utilities-minio`
- S3 API — http://localhost:9000
- веб-консоль — http://localhost:9001 (логин `root`, пароль `rootpassword`)

Учётные данные задаются в `.env` (файл в репозиторий не попадает,
образец — `.env.example`). Данные хранилища лежат в именованном томе
`minio-data`, а не в каталоге проекта.

### Бакет `utilities`

По умолчанию бакеты в MinIO приватные. Бакет создаётся и открывается
на чтение командами:

```bash
docker exec -it utilities-minio mc alias set myminio http://localhost:9000 root rootpassword
docker exec -it utilities-minio mc mb myminio/utilities
docker exec -it utilities-minio mc anonymous set public myminio/utilities
```

Изображения и видео ресурсов загружены с ключами на латинице — на них
ссылаются поля `imageUrl` и `videoUrl`:

| Ресурс | Изображение | Видео |
|---|---|---|
| Электроэнергия | `electricity.jpg` | `electricity.mp4` |
| Отопление | `heating.jpg` | `heating.mp4` |
| Холодное водоснабжение | `cold-water.jpg` | `cold-water.mp4` |
| Газоснабжение | `gas-supply.jpg` | `gas-supply.mp4` |
| Вывоз ТКО | `waste-removal.jpg` | `waste-removal.mp4` |
| Капитальный ремонт | `capital-repair.jpg` | `capital-repair.mp4` |

Пример ссылки: `http://localhost:9000/utilities/electricity.jpg`

Если у услуги нет своего изображения или видео — подставляется заглушка,
которая хранится прямо в коде (`resources/img/default-image.svg`,
`resources/video/default-video.mp4`), а не в MinIO.
