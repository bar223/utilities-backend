# Защита лабораторной работы 4 — порядок показа

## Подготовка

1. `docker compose up -d` — живы `utilities-postgres`, `utilities-minio`,
   `utilities-redis`, `utilities-adminer`.
2. В `.env` есть `JWT_SECRET`, `REDIS_HOST/PORT/PASSWORD` (см. `.env.example`).
3. `go run ./cmd/migrate`, `go run ./cmd/utilities`.
4. Postman: импорт `docs/lab4/postman-collection.json`.
5. Если пользователя нет — запрос «1. POST регистрация».

## Скриншоты 1-3 — Swagger в режиме инкогнито

1. Chrome → новое окно инкогнито (Ctrl+Shift+N) →
   http://localhost:8090/swagger/index.html. Виден список методов, у методов
   создателя — замок.
2. `POST /users/login` → Try it out → `{"login":"newuser","password":"secret"}`
   → Execute: 200 и `accessToken`. Скопировать токен.
3. `GET /utilities` → Try it out → tariff=50 → Execute: 200, список.
   Затем Authorize → `Bearer <токен>` → Authorize; замки закрылись.

## Скриншоты 4-5 — список работает независимо от токена (Postman)

4. «3. GET список» с вкладкой Authorization = Bearer Token (токен есть) →
   200, у своих услуг `isMine: 1`.
5. Тот же запрос с Authorization = No Auth → 200, у всех `isMine: 0`.

## Скриншоты 6-8 — PUT публикации и Redis

6. «4. POST добавить» (с токеном) → 201, затем «6. PUT опубликовать» с
   Authorization = No Auth → **401** `{"error":"требуется авторизация"}`.
   (Перед этим удобно снять черновик «5. GET черновик».)
7. «6. PUT опубликовать» с токеном → **200**, `status: published`, поля
   по теме обновились.
8. «10. POST деавторизация» → 204. Затем в cmd:
   ```
   docker exec -it utilities-redis redis-cli -a password keys "utilities.*"
   docker exec -it utilities-redis redis-cli -a password ttl "utilities.jwt.<токен>"
   ```
   Виден отозванный токен и оставшийся срок его жизни в секундах. После этого
   «5. GET черновик» со старым токеном → 401 «токен отозван».

## Скриншот 9 — StarUML

Sequence-диаграмма: у сообщений свойство `signature` выбрано из операций
доменов диаграммы классов (`docs/diagrams/utilities.mdj`).

## Код для вопросов

- `internal/app/handler/middleware.go` — `WithAuthCheck`, `WithOptionalAuth`,
  проверка подписи и чёрного списка;
- `internal/app/handler/user.go` — выдача JWT (`LoginUser`), logout в Redis;
- `internal/app/redis/redis.go` — ключ `utilities.jwt.<токен>` с TTL;
- `internal/app/handler/utility.go` — `currentUserID(ctx)` вместо константы,
  создатель подставляется из токена;
- `cmd/utilities/main.go` — общие аннотации Swagger и `BearerAuth`.
