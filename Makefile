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

.PHONY: tidy test build run run-with-env docker-build docker-run docker-stop

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
		$(IMAGE)

docker-stop:
	docker stop $(CONTAINER)
