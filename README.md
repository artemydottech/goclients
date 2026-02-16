# goclients

Проект для обучения. Готовая инфраструктура для **selfhosted** разворачивания аналога сервиса [`yclients`](https://www.yclients.com/) с использованием **Go, SQLite**.

Быстрый старт:
Для компиляции go-sqlite3 требуется установка `GCC`.

```
git clone https://github.com/artemydottech/goclients.git
cd goclients
go mod tidy (Установка зависимостей)
```

Тестирование:

```
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Иван"}'
```
