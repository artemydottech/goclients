# goclients

Учебный проект: selfhosted аналог yclients. Go 1.24, `net/http` (ServeMux), SQLite через `mattn/go-sqlite3` (нужен CGO и GCC), конфиг через `.env`.

Проект учебный. Режим работы описан в output style `mentor` (`.claude/output-styles/mentor.md`). Прогресс и роадмап в `LEARNING.md`: прочитай его в начале сессии, обнови в конце.

## Команды

```bash
go run .
gofmt -l .
go vet ./...
go test -race ./...
```

Перед словом "готово" прогнать все четыре проверки из CI: gofmt, build, vet, test -race.

## Структура

- `main.go`: wiring зависимостей, роуты, запуск сервера
- `internal/models`: сущности, `ValidationError`
- `internal/repository`: SQL, схема и миграции в `db.go`
- `internal/service`: бизнес-логика и валидация, тесты рядом
- `internal/handlers`: HTTP, интерфейсы сервисов объявлены здесь

Зависимости направлены внутрь: handlers -> service -> repository. Интерфейсы на стороне потребителя.

## Правила

- Общение на русском, коротко. Никогда не использовать длинное тире.
- Без комментариев в коде (исключение: `TODO(human)` в режиме mentor).
- Новые зависимости только после явного "ставь". По умолчанию стандартная библиотека.
- `git add`, `git commit`, `git push` только по явной просьбе. Перед коммитом показать diff и сообщение.
- Conventional Commits на английском, без trailer Co-Authored-By.
- README обновлять вместе с изменением API.
