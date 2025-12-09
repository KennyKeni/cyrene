# Simple Makefile for a Go project

# Load .env file
include .env
export

# Required for encoding/json/v2 (used by Elysia/openai-go)
export GOEXPERIMENT := jsonv2

# Build the application
all: build test

build:
	@echo "Building..."
	
	
	@go build -o main cmd/api/main.go

# Run the application
run:
	@echo "Running...."


	@go run cmd/api/main.go
# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

test:
	@echo "Running unit tests..."
	@go test ./... -v

itest: test-env-up test-migrate
	@echo "Running integration tests..."
	@set -a && . ./.env.test && set +a && go test ./... -v -tags=integration
	@$(MAKE) test-env-down

test-env-up:
	@unset DB_PORT DB_DATABASE QDRANT_PORT KAFKA_PORT && \
		docker compose -p cyrene-test --env-file .env.test -f docker-compose.test.yml up -d --wait

test-env-down:
	@unset DB_PORT DB_DATABASE QDRANT_PORT KAFKA_PORT && \
		docker compose -p cyrene-test --env-file .env.test -f docker-compose.test.yml down

test-migrate:
	@set -a && . ./.env.test && set +a && \
		goose -dir migrations postgres "postgres://$$DB_USERNAME:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_DATABASE?sslmode=disable" up

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Run database migrations
migrate-up:
	@goose -dir migrations postgres "postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=disable" up

migrate-down:
	@goose -dir migrations postgres "postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=disable" down

# Generate Jet models from database schema
jet-gen:
	@jet -dsn="postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=disable" -schema=$(DB_SCHEMA) -path=./internal/platform/postgres/jet

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test clean watch docker-run docker-down itest test-env-up test-env-down test-migrate migrate-up migrate-down jet-gen
