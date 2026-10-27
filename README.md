# Utilities — backend

Учебный проект по дисциплине **«Разработка Интернет Приложений» (РИП), 2026**.

Тема варианта — **«Квитанция по квартплате»**. Предметная область: начисления за
коммунальные ресурсы, из которых складывается месячная квитанция жильца.

Фронтенд — репозиторий [utility-bills-frontend](https://github.com/bar223/utility-bills-frontend).

## Лабораторная работа 3 — веб-сервис (REST API)

Ветка: `lab3-rest-api`

Тема лабораторной — веб-сервис со всей бизнес-логикой (кроме
авторизации) для будущего SPA. Сервис отдаёт только JSON и тестируется
через Postman — HTML-страниц и шаблонов в этой ветке нет (интерфейс
лаб 1-2 остался в ветке `lab2-database`, новый появится во фронтенде,
лабораторные 5-7).

- ровно 10 HTTP-методов: 7 в домене услуг и 3 в домене пользователей;
  все адреса начинаются с `/api`;
- в теле ответа нет поля `status`/`success` — результат передаётся только
  кодом состояния HTTP (200, 201, 204, 400, 404, 409, 500). Тело
  успешного ответа — сам объект (или массив), тело ошибки — `{"error": "…"}`;
- взаимодействие с БД — только через ORM (GORM);
- записи в статусе `deleted` клиенту не отдаются;
- статусы меняются только в одну сторону, двумя разными методами
  создателя: PUT (`draft → published`) и DELETE (`→ deleted`);
- системные поля (`id`, `status`, создатель, даты) с клиента не
  принимаются — их вычисляет бэкенд;
- пользователь-создатель зафиксирован константой через функцию-синглтон
  `handler.CurrentUserID()` (`internal/app/handler/handler.go`);
- коллекция Postman — `docs/lab3/postman-collection.json` (10 запросов).

### Методы API

**Домен услуг** — `/api/utilities`

| № | Метод | URL | Тело запроса | Ответ |
|---|---|---|---|---|
| 1 | GET | `/api/utilities?tariff=10` | — | 200, массив опубликованных услуг с тарифом не выше `tariff` (без параметра — все). У каждой `isMine`: 1 — создатель услуги = текущий пользователь, 0 — нет; `likesCount` |
| 2 | GET | `/api/utilities/feed?id=1&next=true` | — | 200, одна опубликованная услуга с вложенным `creator {id, login}`, `liked`: 1 — текущий пользователь лайкнул, 0 — нет. Без `id` — первая услуга (вход из панели вкладок), с `id` — эта услуга, с `id` и `next=true` — следующая (после последней — снова первая). 404 — нет такой |
| 3 | GET | `/api/utilities/draft` | — | 200, черновик текущего пользователя (не более одного, id не указывается); 404 — черновика нет |
| 4 | POST | `/api/utilities` | `multipart/form-data`: `name` (обязательно), `image`, `video` — файлы | 201, созданный черновик. Файлы загружаются в MinIO под латинскими именами `utility-<id>-image.<ext>` / `utility-<id>-video.<ext>`, ссылки сохраняются в `image_url` / `video_url`. 400 — не тот тип файла, 409 — черновик уже есть |
| 5 | PUT | `/api/utilities/:id` | JSON `{description, tariffRate, consumptionNorm, measurementUnit}` | 200, опубликованная услуга (статус `draft → published`, `publishedAt` = сейчас); 404 — это не черновик текущего пользователя |
| 6 | DELETE | `/api/utilities/:id` | — | 204 без тела, `status → deleted` (soft delete); 404 — нет такой услуги у текущего пользователя |
| 7 | POST | `/api/utilities/:id/like` | JSON `{"value": 1}` — 1 ставит, 0 снимает лайк | 200, `{liked, likesCount}`; 404 — услуга не опубликована |

**Домен пользователей** — `/api/users`

| № | Метод | URL | Тело запроса | Ответ |
|---|---|---|---|---|
| 8 | POST | `/api/users/register` | JSON `{login, password}` | 201, `{id, login, isModerator}`; 409 — логин занят |
| 9 | POST | `/api/users/login` | — | 200, заглушка — аутентификация в лабораторной 4 |
| 10 | POST | `/api/users/logout` | — | 200, заглушка — деавторизация в лабораторной 4 |

### Таблицы БД

N — необязательное поле (NULL), остальные NOT NULL.

**utilities** — услуги

| Поле | Тип | Описание |
|---|---|---|
| `id` | BIGINT, PK | идентификатор |
| `name` | VARCHAR(100) | наименование |
| `description` | VARCHAR(500), N | краткое описание |
| `status` | VARCHAR(15) | `draft` / `published` / `deleted` |
| `image_url` | VARCHAR(255) | ссылка на изображение в MinIO |
| `video_url` | VARCHAR(255) | ссылка на видео в MinIO |
| `tariff_rate` | NUMERIC(10,2), N | тариф за единицу — поле по теме, по нему фильтр |
| `consumption_norm` | NUMERIC(10,3), N | норматив потребления — поле по теме |
| `measurement_unit` | VARCHAR(20) | единица измерения, `DEFAULT ''` до публикации |
| `created_at` | TIMESTAMP | дата создания |
| `published_at` | TIMESTAMP, N | дата формирования (публикации) |
| `creator_id` | BIGINT, FK → users.id | создатель |

**likes** — лайки, м-м «пользователь — услуга»

| Поле | Тип | Описание |
|---|---|---|
| `id` | BIGINT, PK | идентификатор |
| `user_id` | BIGINT, FK → users.id | кто поставил |
| `utility_id` | BIGINT, FK → utilities.id | какой услуге; пара (`user_id`, `utility_id`) уникальна |

**users** — пользователи

| Поле | Тип | Описание |
|---|---|---|
| `id` | BIGINT, PK | идентификатор |
| `login` | VARCHAR(50), U | логин, уникальный |
| `password` | VARCHAR(100) | пароль (в JSON не отдаётся) |
| `is_moderator` | BOOLEAN | признак модератора |

Каскадного удаления нет, удаление услуг только логическое (`status`).

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

## Структура проекта (ветка lab3-rest-api)

```
cmd/
  utilities/main.go          точка входа веб-сервиса
  migrate/main.go            миграции (AutoMigrate по моделям)
config/config.toml           хост и порт сервиса (8090)
internal/
  pkg/app.go                 сборка приложения и запуск HTTP-сервера
  app/
    config/                  чтение config.toml
    dsn/                     строка подключения к PostgreSQL из .env
    ds/
      utility.go             модель Utility (таблица utilities)
      like.go                модель Like (таблица likes)
      user.go                модель User (таблица users)
      serializers.go         сериализаторы ответов и тел запросов
    handler/
      handler.go             маршруты /api, CurrentUserID(), ошибки
      utility.go             7 методов домена услуг
      user.go                3 метода домена пользователей
    repository/
      repository.go          подключение к PostgreSQL и MinIO
      env.go                 настройки из .env
      utility.go             запросы к utilities через GORM
      like.go                запросы к likes
      user.go                запросы к users
      upload.go              загрузка файлов в MinIO
docs/
  lab2/                      скриншоты и ER-спецификация ЛР-2
  lab3/                      коллекция Postman, сценарий защиты
  diagrams/utilities.mdj     StarUML: ER-диаграмма и диаграмма классов
docker-compose.yml           MinIO, PostgreSQL, Adminer
```

## Запуск

```bash
cp .env.example .env
docker compose up -d        # MinIO, PostgreSQL, Adminer
go run ./cmd/migrate        # схема БД
go run ./cmd/utilities      # веб-сервис на http://localhost:8090/api
```

Postman: **File → Import** → `docs/lab3/postman-collection.json`.

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

Если у услуги нет своего изображения или видео, веб-сервис отдаёт пустую
ссылку. На HTML-страницах ЛР-2 (ветка `lab2-database`) вместо неё
показывается заглушка, которая хранится на SSR-сервере
(`resources/img/default-image.svg`, `resources/video/default-video.mp4`),
а не в MinIO; в SPA заглушки будут храниться во фронтенде.
