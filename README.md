# Сервис поиска и бронирования парковочных мест

Учебный проект по теме **Twelve-Factor App** на Go.

## Что уже реализовано

- REST API для фронтенда:
  - `GET /health`
  - `GET /api/v1/spots`
  - `POST /api/v1/bookings`
- Бизнес-логика поиска доступных мест и бронирования.
- In-memory хранилище для локальной разработки и тестов.
- Unit-тесты для ключевых сценариев.

## Фактор I. Codebase

**One codebase tracked in revision control, many deploys.**

Это реализовано так:

- Один репозиторий `parking-service` для одного сервиса.
- Один модуль Go (`go.mod`) и единый entrypoint приложения: `cmd/parking-api`.
- Выделенные внутренние модули (`internal/parking`) и версия REST API (`api/v1`).
- Одинаковый код может быть развернут в разных окружениях (dev/stage/prod) без изменений исходников.

## Фактор II. Dependencies

**Explicitly declare and isolate dependencies.**

Это реализовано так:

- Все зависимости объявлены через Go Modules (`go.mod`, `go.sum`).
- Зависимости фиксируются командой `go mod tidy`.
- Сборка/тесты выполняются с `-mod=readonly`, чтобы исключить неявные изменения зависимостей.
- Сервис не использует `os/exec` и не зависит от внешних системных утилит вроде `curl`.
- Для контейнерного деплоя добавлен `Dockerfile` с воспроизводимой сборкой бинаря.

## Структура проекта

```text
parking-service/
  api/v1/                # REST handlers (версия API)
  cmd/parking-api/       # main.go (точка входа)
  internal/parking/      # доменная модель, сервис, репозиторий, тесты
  go.mod
  go.sum
  Makefile
  Dockerfile
```

## Быстрый старт

```bash
go mod tidy
go test ./...
go run -mod=readonly ./cmd/parking-api
```

Сервис стартует на порту `8080` (или из переменной `HTTP_PORT`).

## Примеры запросов

Получить список мест:

```bash
curl "http://localhost:8080/api/v1/spots?location=center&vehicleType=car&maxPrice=180"
```

Забронировать место:

```bash
curl -X POST "http://localhost:8080/api/v1/bookings" \
  -H "Content-Type: application/json" \
  -d '{
    "spotId": "A-101",
    "userId": "student-42",
    "from": "2026-03-20T10:00:00Z",
    "to": "2026-03-20T12:30:00Z"
  }'
```

## Что дальше по курсу

На следующих практиках можно последовательно добавлять следующие факторы:

- конфигурацию через environment variables;
- backing services (PostgreSQL, Redis) как attachable resources;
- отдельные build/release/run стадии;
- stateless-процессы и horizontal scaling;
- gRPC-контракты для межсервисного взаимодействия при сохранении REST наружу.
