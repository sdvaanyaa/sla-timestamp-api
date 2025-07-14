include .env.example
export

BINARY_NAME = sla-timestamp-api
BUILD_DIR = build
MIGRATIONS_DIR = migrations
DATABASE_DSN = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)

.PHONY: all update linter build start run clean bin-deps up down restart goose-add goose-up goose-down goose-status test test-coverage mock

all: run

update:
	@echo "Updating dependencies"
	@go mod tidy

linter:
	@echo "Running linters"
	@golangci-lint run ./... --tests=false

build:
	@echo "Building application"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/sla-timestamp-api/main.go

start:
	@echo "Starting application"
	@$(BUILD_DIR)/$(BINARY_NAME)

run: bin-deps up goose-up update linter build start

clean:
	@echo "Cleaning up"
	@rm -rf $(BUILD_DIR)
	@go clean

bin-deps:
	@echo "Installing goose dependencies"
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/gojuno/minimock/v3/cmd/minimock@latest

swagger:
	@echo "Generating Swagger docs"
	@swag init --dir ./cmd/sla-timestamp-api,./internal/handler,./internal/entity --generalInfo main.go --output ./docs

swagger-fmt:
	@echo "Formatting Swagger comments"
	@swag fmt

up:
	@echo "Starting Docker Compose"
	@docker-compose up -d

down:
	@echo "Stopping Docker Compose"
	@docker-compose down

restart: down up

goose-add:
	@echo "Creating new migration"
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_DSN)" create $(NAME) sql

goose-up:
	@echo "Waiting for Postgres to be ready"
		@for i in {1..30}; do \
			pg_isready -h $(POSTGRES_HOST) -p $(POSTGRES_PORT) -U $(POSTGRES_USER) && break; \
			echo "Postgres not ready, retry $$i/30"; \
			sleep 1; \
		done
	@echo "Applying migrations"
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_DSN)" up

goose-down:
	@echo "Reverting migrations"
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_DSN)" down

goose-status:
	@echo "Checking migration status"
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_DSN)" status

test:
	@echo "Running tests"
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage"
	@go test -v -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

mock:
	@echo "Generating mocks for interfaces"
	@mkdir -p internal/repository/mocks
	@minimock -i github.com/sdvaanyaa/sla-timestamp-api/internal/repository.TimestampStorage -o internal/repository/mocks/repository_mock.go
	@mkdir -p pkg/cache/mocks
	@minimock -i github.com/sdvaanyaa/sla-timestamp-api/pkg/cache.Cache -o pkg/cache/mocks/cache_mock.go