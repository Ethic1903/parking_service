# Parking Service

Учебный сервис поиска и бронирования парковочных мест на Go.

## Что реализовано в этой практике

- Приложение переведено на внешнее хранилище состояния (PostgreSQL/SQLite).
- In-memory хранение удалено из runtime-кода.
- Добавлен SQL-репозиторий с транзакционным бронированием.
- Добавлен `docker-compose.yml` с PostgreSQL как attachable resource.
- Все параметры подключения к БД читаются из env (через `viper`).
- Поддержан локальный режим с SQLite и контейнерный режим с PostgreSQL.

## API

- `GET /health`
- `GET /api/v1/spots`
- `POST /api/v1/bookings`

## Конфигурация (env)

HTTP:

- `APP_ENV`
- `HTTP_PORT`
- `HTTP_READ_TIMEOUT_SEC`
- `HTTP_WRITE_TIMEOUT_SEC`
- `HTTP_IDLE_TIMEOUT_SEC`
- `HTTP_SHUTDOWN_TIMEOUT_SEC`

Storage:

- `DB_DRIVER` (`sqlite` или `postgres`)
- `DB_SQLITE_PATH` (для `sqlite`)
- `DB_POSTGRES_HOST`
- `DB_POSTGRES_PORT`
- `DB_POSTGRES_DBNAME`
- `DB_POSTGRES_USER`
- `DB_POSTGRES_PASSWORD`
- `DB_POSTGRES_SSLMODE`

Дополнительно:

- `CONFIG_FILE` (опциональный путь к YAML)

Пример значений есть в `.env.example`.

## Локальный запуск (SQLite)

```bash
go mod tidy
go test ./...
DB_DRIVER=sqlite DB_SQLITE_PATH=parking.db go run ./cmd/parking-api
```

## Docker Compose (PostgreSQL)

```bash
docker compose up -d --build
docker compose ps
```

Сервис `app` подключается к БД по имени сервиса `postgres` внутри compose-сети.

## Выполнение шагов практики 1-5

### 1) Добавление БД через Docker Compose

Реализовано в `docker-compose.yml`:

- сервис `postgres`;
- сервис `app`;
- общая сеть `parking-net`;
- volume `postgres-data` для сохранения данных.

### 2) Подключение к БД через переменные окружения

Реализовано в `tools/config` и `tools/storage/db.go`:

- `DB_DRIVER=sqlite` для локальной разработки;
- `DB_DRIVER=postgres` в контейнере;
- строка подключения строится из env-переменных.

### 3) Перенос состояния из памяти в БД

Реализовано в `internal/pkg/repository/sql_repository.go` и `migrations/`:

- места и их доступность хранятся в таблице `parking_spots`;
- бронирования хранятся в таблице `bookings`;
- бронирование выполняется в транзакции (`UPDATE availability + INSERT booking`).

### 4) Симуляция сбоя процесса

Запуск нескольких экземпляров:

```bash
docker compose up -d --scale app=3
docker compose ps
```

Остановите один контейнер app:

```bash
docker stop <app_container_name>
```

Проверьте, что другие экземпляры продолжают работать, а записи не теряются, так как хранятся в PostgreSQL.

### 5) Проверка stateless

```bash
docker compose stop app
docker compose up -d app
```

Или полный рестарт:

```bash
docker compose down
docker compose up -d
```

Данные сохраняются благодаря `postgres-data` volume и внешней БД.

## Примеры запросов

```bash
curl "http://localhost:8080/api/v1/spots?location=center&vehicleType=car&maxPrice=180"
```

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
