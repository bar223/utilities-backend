# Диаграммы StarUML

`utilities.mdj` — один файл StarUML со всеми диаграммами:

- **ER Diagram** (Data Model) — таблицы `utilities`, `users`, `likes`,
  компактно «квадратом»: `utilities` слева, `users` и `likes` справа.
  N — необязательные поля, их ровно 4: `description`, `tariff_rate`,
  `consumption_norm`, `published_at`; `image_url`, `video_url` обязательные.
- **Диаграмма классов** (Model) — 4 страницы фронтенда по теме (названия
  переносятся), интерфейсы `API utilities` и `API users` в виде
  «фонарика» (один провод от домена, лучи к страницам), домены с 10
  методами по url, модели `Utility`, `Like`, `User`, таблицы БД.

Перед экспортом в отчёт: View → Show Grid выключить, File → Export
Diagram As → PNG; лист А4.
