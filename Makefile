GO ?= go
APP ?= parking-api
CMD ?= ./cmd/parking-api
GOFLAGS ?= -mod=readonly

.PHONY: tidy test build run

tidy:
	$(GO) mod tidy

test:
	$(GO) test $(GOFLAGS) ./...

build:
	$(GO) build $(GOFLAGS) -o ./bin/$(APP) $(CMD)

run:
	$(GO) run $(GOFLAGS) $(CMD)
