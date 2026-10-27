# Защита лабораторной работы 3 — порядок показа

## Подготовка

1. `docker compose up -d` — живы `utilities-postgres`, `utilities-minio`,
   `utilities-adminer`.
2. `go run ./cmd/migrate`, затем `go run ./cmd/utilities` — сервис на
   `http://localhost:8090/api`. Браузер не нужен: HTML-страниц в ЛР-3 нет.
3. Postman: **File → Import** → `docs/lab3/postman-collection.json`.
4. Черновика у пользователя 1 быть не должно (иначе запрос 4 вернёт 409):
   Adminer → `SELECT id FROM utilities WHERE status = 'draft';`.

Скриншоты Postman делать с масштабом 125-150 % (View → Zoom In) и
обрезать до адреса, вкладки Body/Params и ответа с кодом — так текст
читается в документе.

## Скриншот 1 — коллекция

Левая панель Postman с раскрытой коллекцией: ровно 10 запросов.

## Скриншоты 2-10 — запросы (код ответа виден справа над телом)

| Скрин | Запрос коллекции | Что показать |
|---|---|---|
| 2 | 1. GET список (`?tariff=10`) | 200, массив, у каждой `isMine` 0/1 |
| 3 | 2. POST добавить | вкладка Body: `name`, файлы `image`, `video`; ответ 201, `imageUrl`/`videoUrl` с латинскими именами. id сам запишется в переменную `draftId` |
| 4 | 3. GET черновик | 200, тот же id, `status: draft` |
| 5 | 4. PUT опубликовать | 200, `status: published`, заполнены поля по теме |
| 6 | 5. GET лента — галочки `id`, `next` сняты | 200, первая услуга, `liked` 0/1, вложенный `creator` |
| 7 | 5. GET лента — `id=1`, `next=true` | 200, следующая услуга после id=1 |
| 8 | 6. POST лайк `{"value": 1}` | 200, `liked: 1`, `likesCount` вырос |
| 9 | 7. DELETE удалить | 204 No Content (тела нет — результат в коде) |
| 10 | 8. POST регистрация | 201, `{id, login, isModerator}`, пароля в ответе нет |

## Скриншоты 11-13 — select в БД (Adminer → SQL-запрос)

```sql
SELECT id, name, status, image_url, video_url, published_at FROM utilities ORDER BY id;
SELECT * FROM likes ORDER BY id;
SELECT id, login, is_moderator FROM users ORDER BY id;
```

Видно: добавленная услуга в `deleted`, лайк в `likes`, новый пользователь.

## Скриншоты 14-15 — модели и сериализаторы

- `internal/app/ds/utility.go` — модель `Utility` с тегами `gorm` и `json`;
- `internal/app/ds/serializers.go` — `UtilityListItem` (`isMine`),
  `UtilityFeedItem` (`liked`, вложенный `creator` — вложенная
  сериализация), тела запросов `PublishUtilityRequest`, `LikeRequest`
  (в них нет системных полей).

## Скриншоты 16-17 — синглтон и README

- `internal/app/handler/handler.go` — константа `currentUserID` и
  функция-синглтон `CurrentUserID()`; там же 10 маршрутов `/api`;
- `internal/app/handler/utility.go` — вызовы `CurrentUserID()` в методах;
- `README.md` — таблица 10 методов и таблицы БД с полями.

## Ответы на вопросы

- **Почему лента работает без параметров?** Так требует методичка: при
  переходе в ленту из панели вкладок id не указывается — отдаётся первая
  опубликованная услуга. С `id` — эта услуга, с `id` и `next=true` —
  следующая. Это один метод с необязательными параметрами.
- **Где success?** Нигде: по лекции 5 статус передаётся только кодом
  HTTP (200/201/204/400/404/409/500), в теле его нет.
- **Почему нет интерфейса?** ЛР-3 — веб-сервис для SPA, проверяется через
  Postman. HTML-страницы лаб 1-2 остались в ветке `lab2-database`.
