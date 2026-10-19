# LEARNING

Журнал обучения на проекте goclients. Агент читает его в начале сессии и обновляет в конце.

Цель: быть готовым к production-разработке на Go. Уметь самостоятельно спроектировать, написать, протестировать, задеплоить и сопровождать HTTP-сервис с БД.

## Текущий этап

Этап 1, шаг не начат.

## Уже освоено

- Слои handlers / service / repository, интерфейсы на стороне потребителя, DI через конструкторы
- Ошибки: sentinel, обертка через `%w`, `errors.Is`, `errors.As`, свой тип `ValidationError`
- `net/http`: ServeMux с методами и path-параметрами, JSON encode/decode, таймауты сервера
- SQLite: схема, индексы, внешние ключи, транзакции для защиты от двойного бронирования
- Юнит-тесты сервисов, CI: gofmt, build, vet, test -race
- Часовые пояса, статусы записей, снапшот цены

## Типичные ошибки (проверять на каждом ревью)

- Результат функции не присвоен переменной
- `err` переобъявлен через `:=` вместо `=`
- Нет `return` после ветки обработки ошибки
- В `%w` обернута не та переменная
- `fmt.Println` вместо stderr или логгера

## Роадмап к production

Состояние кода на 2026-09-21, от которого строится план: `context` не используется нигде; graceful shutdown нет; `database/sql` импортируется в handlers и service; тестов хендлеров (`httptest`) нет; логирование через `log`; схема создается на старте без версионных миграций; цена хранится как REAL; нет авторизации, пагинации, health-эндпоинта, Dockerfile и линтера.

1. **Context и жизненный цикл.** Прокинуть `context.Context` от хендлера до запроса в БД (`QueryContext`, `ExecContext`, `BeginTx`). Graceful shutdown по SIGINT/SIGTERM через `signal.NotifyContext` и `server.Shutdown`.
2. **Границы слоев и ошибки.** Убрать `database/sql` из handlers и service: доменные ошибки (`ErrNotFound`, `ErrConflict`) в models, маппинг `sql.ErrNoRows` в repository, единая функция перевода ошибки в HTTP-статус.
3. **Middleware.** Recover, request id, access-лог на `log/slog`, лимит размера тела. Единый JSON-формат ошибок вместо `http.Error` с текстом.
4. **Тесты HTTP-слоя.** `httptest`, table-driven тесты хендлеров с фейковым сервисом, тесты repository на временной БД, покрытие и `-race`.
5. **Конфигурация.** Структура Config, чтение и валидация env на старте, падение с понятной ошибкой.
6. **Миграции и деньги.** Версионные миграции вместо схемы на старте. Перевод цены из REAL в целые копейки с миграцией данных.
7. **PostgreSQL.** Переезд с SQLite на Postgres, пул соединений, docker-compose для локальной разработки, интеграционные тесты на реальной БД, уровни изоляции и блокировки для бронирования.
8. **Аутентификация и авторизация.** Хеширование паролей, сессии или JWT, доступ только к данным своей компании (multi-tenant), middleware авторизации.
9. **API-гигиена.** Пагинация, фильтры, PATCH/PUT для обновления сущностей, идемпотентность создания записи, OpenAPI-описание.
10. **Наблюдаемость.** `/healthz` и `/readyz`, метрики Prometheus, структурные логи с request id, pprof.
11. **Конкурентность на практике.** Фоновый воркер напоминаний о записи: горутины, каналы, `errgroup`, тикеры, корректная остановка по context.
12. **Поставка.** Multi-stage Dockerfile, golangci-lint в CI, сборка образа в CI, деплой на VPS или Railway, бэкапы БД, откат релиза.

Порядок можно менять. Каждый этап закрывается, когда Артемий может объяснить сделанное без подсказок.

## Журнал сессий

Формат записи: дата, что сделано, что осталось непонятым, следующий шаг.

## Сделано агентом, разобрать позже


- 2026-10-08, этап 2. `models.ErrNotFound` вместо `sql.ErrNoRows`: repository превращает `sql.ErrNoRows` в доменную ошибку, service и handlers проверяют `errors.Is(err, models.ErrNotFound)`. `database/sql` остался только в repository. Разобрать: почему `errNoRows` склеен из двух ошибок через `%w: %w`.
- 2026-10-09, этап 3. Единый JSON-формат ошибок: `writeError`, `writeStatusJSON`, `writeJSON` в `internal/handlers/respond.go`. Разобрать: чем `http.Error` отличается от своего хелпера и почему Content-Type надо ставить до `WriteHeader`.
- 2026-10-10, этап 3 и 5. Пакет `internal/middleware` (Recover, RequestID, AccessLog на `log/slog`, BodyLimit, SecureHeaders, Chain), пакет `internal/config` с валидацией env, `/healthz` и `/readyz`, `decodeJSON` с ответом 413. Разобрать: порядок middleware в `Chain` и почему RequestID стоит перед Recover.
- 2026-10-11, этап 4. Хендлеры покрыты `httptest` с фейковыми сервисами: users, companies, employees, services, clients. `slog` в хендлерах берет request id из контекста через `middleware.NewContextHandler`. Разобрать: как устроен `slog.Handler` и зачем `WithAttrs`/`WithGroup` в обертке.
- 2026-10-12, этап 4. Тесты хендлеров assignments, schedule, slots, time-off, appointments и тесты `models`. Тест всей цепочки middleware нашел ошибку порядка: паника логировалась без request id. Разобрать: как это воспроизвести без теста.
- 2026-10-13. `pathID` вместо повторяющегося разбора `{id}` в каждом хендлере.
- 2026-10-14..2026-10-17, этап 1. `context.Context` проброшен от хендлера до запроса в БД: repository использует `ExecContext`, `QueryContext`, `QueryRowContext`, `BeginTx`, service и handlers принимают `ctx` первым параметром, `Open` и `Migrate` тоже. Разобрать: почему ctx первый параметр и почему его не хранят в структуре.
- 2026-10-18, этап 4. Тесты repository: отмененный контекст прерывает запрос и возвращает `context.Canceled`.
- 2026-10-19, этап 4. Тесты хендлеров: сервис получает контекст запроса (`serveWithContext`, `assertMarked`).

Не сделано из роадмапа: версионные миграции и копейки (6), Postgres (7), авторизация (8), пагинация (9), метрики и pprof (10), воркер напоминаний (11), Dockerfile и линтер (12).
