GO ?= go
APP ?= parking-api
CMD ?= ./cmd/parking-api
GOFLAGS ?= -mod=readonly
IMAGE ?= parking-api:local
CONTAINER ?= parking-api-local
APP_ENV ?= dev
HTTP_PORT ?= 8080
HTTP_READ_TIMEOUT_SEC ?= 5
HTTP_WRITE_TIMEOUT_SEC ?= 10
HTTP_IDLE_TIMEOUT_SEC ?= 60
HTTP_SHUTDOWN_TIMEOUT_SEC ?= 10
DB_DRIVER ?= sqlite
DB_SQLITE_PATH ?= parking.db
DB_POSTGRES_HOST ?= localhost
DB_POSTGRES_PORT ?= 5432
DB_POSTGRES_DBNAME ?= parking
DB_POSTGRES_USER ?= parking
DB_POSTGRES_PASSWORD ?= parking
DB_POSTGRES_SSLMODE ?= disable

.PHONY: tidy test build run run-with-env docker-build docker-run docker-stop compose-up compose-down compose-scale verify-stateless

tidy:
	$(GO) mod tidy

test:
	$(GO) test $(GOFLAGS) ./...

build:
	$(GO) build $(GOFLAGS) -o ./bin/$(APP) $(CMD)

run:
	$(GO) run $(GOFLAGS) $(CMD)

run-with-env:
	APP_ENV=$(APP_ENV) \
	HTTP_PORT=$(HTTP_PORT) \
	HTTP_READ_TIMEOUT_SEC=$(HTTP_READ_TIMEOUT_SEC) \
	HTTP_WRITE_TIMEOUT_SEC=$(HTTP_WRITE_TIMEOUT_SEC) \
	HTTP_IDLE_TIMEOUT_SEC=$(HTTP_IDLE_TIMEOUT_SEC) \
	HTTP_SHUTDOWN_TIMEOUT_SEC=$(HTTP_SHUTDOWN_TIMEOUT_SEC) \
	DB_DRIVER=$(DB_DRIVER) \
	DB_SQLITE_PATH=$(DB_SQLITE_PATH) \
	DB_POSTGRES_HOST=$(DB_POSTGRES_HOST) \
	DB_POSTGRES_PORT=$(DB_POSTGRES_PORT) \
	DB_POSTGRES_DBNAME=$(DB_POSTGRES_DBNAME) \
	DB_POSTGRES_USER=$(DB_POSTGRES_USER) \
	DB_POSTGRES_PASSWORD=$(DB_POSTGRES_PASSWORD) \
	DB_POSTGRES_SSLMODE=$(DB_POSTGRES_SSLMODE) \
	$(GO) run $(GOFLAGS) $(CMD)

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm --name $(CONTAINER) -p $(HTTP_PORT):$(HTTP_PORT) \
		-e APP_ENV=$(APP_ENV) \
		-e HTTP_PORT=$(HTTP_PORT) \
		-e HTTP_READ_TIMEOUT_SEC=$(HTTP_READ_TIMEOUT_SEC) \
		-e HTTP_WRITE_TIMEOUT_SEC=$(HTTP_WRITE_TIMEOUT_SEC) \
		-e HTTP_IDLE_TIMEOUT_SEC=$(HTTP_IDLE_TIMEOUT_SEC) \
		-e HTTP_SHUTDOWN_TIMEOUT_SEC=$(HTTP_SHUTDOWN_TIMEOUT_SEC) \
		-e DB_DRIVER=$(DB_DRIVER) \
		-e DB_SQLITE_PATH=$(DB_SQLITE_PATH) \
		-e DB_POSTGRES_HOST=$(DB_POSTGRES_HOST) \
		-e DB_POSTGRES_PORT=$(DB_POSTGRES_PORT) \
		-e DB_POSTGRES_DBNAME=$(DB_POSTGRES_DBNAME) \
		-e DB_POSTGRES_USER=$(DB_POSTGRES_USER) \
		-e DB_POSTGRES_PASSWORD=$(DB_POSTGRES_PASSWORD) \
		-e DB_POSTGRES_SSLMODE=$(DB_POSTGRES_SSLMODE) \
		$(IMAGE)

docker-stop:
	docker stop $(CONTAINER)

compose-up:
	docker compose up -d --build

compose-scale:
	docker compose up -d --scale app=3

compose-down:
	docker compose down

verify-stateless:
	@set -eu; \
		docker compose up -d --build; \
		docker compose up -d --scale app=3; \
		PORT1=$$(docker compose port --index 1 app 8080 | sed 's/.*://'); \
		PORT2=$$(docker compose port --index 2 app 8080 | sed 's/.*://'); \
		PORT3=$$(docker compose port --index 3 app 8080 | sed 's/.*://'); \
		BASE1="http://localhost:$$PORT1"; \
		BASE2="http://localhost:$$PORT2"; \
		BASE3="http://localhost:$$PORT3"; \
		USER_ID="verify-$$(date +%s)"; \
		curl -sS -X POST "$$BASE1/api/v1/bookings" \
			-H "Content-Type: application/json" \
			-d "{\"spotId\":\"A-101\",\"userId\":\"$$USER_ID\",\"from\":\"2026-04-03T10:00:00Z\",\"to\":\"2026-04-03T12:30:00Z\"}" >/dev/null || true; \
		EXPECTED=$$(curl -fsS "$$BASE1/api/v1/spots?location=center&vehicleType=car&maxPrice=200"); \
		S2=$$(curl -fsS "$$BASE2/api/v1/spots?location=center&vehicleType=car&maxPrice=200"); \
		S3=$$(curl -fsS "$$BASE3/api/v1/spots?location=center&vehicleType=car&maxPrice=200"); \
		test "$$EXPECTED" = "$$S2"; \
		test "$$EXPECTED" = "$$S3"; \
		docker stop parking-service-app-2 >/dev/null; \
		S1_AFTER_STOP=$$(curl -fsS "$$BASE1/api/v1/spots?location=center&vehicleType=car&maxPrice=200"); \
		S3_AFTER_STOP=$$(curl -fsS "$$BASE3/api/v1/spots?location=center&vehicleType=car&maxPrice=200"); \
		test "$$EXPECTED" = "$$S1_AFTER_STOP"; \
		test "$$EXPECTED" = "$$S3_AFTER_STOP"; \
		docker compose stop app; \
		docker compose up -d app; \
		PORT_RESTART=$$(docker compose port app 8080 | sed 's/.*://'); \
		RESTARTED=$$(curl -fsS "http://localhost:$$PORT_RESTART/api/v1/spots?location=center&vehicleType=car&maxPrice=200"); \
		test "$$EXPECTED" = "$$RESTARTED"; \
		docker compose up -d --scale app=3; \
		PORT1=$$(docker compose port --index 1 app 8080 | sed 's/.*://'); \
		PORT2=$$(docker compose port --index 2 app 8080 | sed 's/.*://'); \
		PORT3=$$(docker compose port --index 3 app 8080 | sed 's/.*://'); \
		F1=$$(curl -fsS "http://localhost:$$PORT1/api/v1/spots?location=center&vehicleType=car&maxPrice=200"); \
		F2=$$(curl -fsS "http://localhost:$$PORT2/api/v1/spots?location=center&vehicleType=car&maxPrice=200"); \
		F3=$$(curl -fsS "http://localhost:$$PORT3/api/v1/spots?location=center&vehicleType=car&maxPrice=200"); \
		test "$$EXPECTED" = "$$F1"; \
		test "$$EXPECTED" = "$$F2"; \
		test "$$EXPECTED" = "$$F3"; \
		echo "verify-stateless: OK"
