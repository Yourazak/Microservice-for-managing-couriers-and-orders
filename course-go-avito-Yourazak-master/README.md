# domashka1-3

## Запуск

1. Создать `.env` в корне (пример в репозитории).
2. Прогнать миграции (goose) или вручную создать таблицу.
3. Запустить:
   go run ./cmd/main.go --port 8080

## Миграции

Используется goose. Переменные из .env должны быть доступны:
POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, POSTGRES_HOST, POSTGRES_PORT

Пример:
DB_DSN=postgres://myuser:mypassword@localhost:5432/test_db?sslmode=disable.


ТЕСТЫ 
# Все основные тесты
go test -v ./internal/handler ./internal/service ./internal/model ./pkg/db -cover

# Интеграционные тесты (требует Docker)
go test -v ./internal/repository -tags=integration

# Coverage отчет
go test -coverprofile=coverage.out ./internal/handler ./internal/service ./internal/model ./pkg/db
go tool cover -func=coverage.out

(make all,make coverage)
