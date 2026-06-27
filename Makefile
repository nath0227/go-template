APP     := service-name
BIN_DIR := bin
BINARY  := $(BIN_DIR)/$(APP)

.PHONY: build run tidy lint test docker-up docker-down migrate

build:
	go build -o $(BINARY) ./cmd/server

run: build
	./$(BINARY)

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

test:
	go test -race ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	migrate -path migrations -database "$${DB_DSN}" up
