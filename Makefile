.PHONY: help build-frontend build-server build-microchat build dev run clean docker-build docker-up docker-down install lint test generate-client

help:
	@echo "Available commands:"
	@echo "  make install           - Install Go and frontend dependencies"
	@echo "  make build             - Build everything (frontend + server + microchat)"
	@echo "  make build-frontend    - Build frontend only"
	@echo "  make build-server      - Build server only"
	@echo "  make build-microchat   - Build microchat client (TUI + CLI) only"
	@echo "  make test              - Run all tests"
	@echo "  make lint              - Run golangci-lint"
	@echo "  make dev               - Run in development mode"
	@echo "  make run               - Build and run server"
	@echo "  make docker-build      - Build Docker image"
	@echo "  make docker-up         - Start with docker-compose"
	@echo "  make docker-down       - Stop docker-compose"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make generate-client   - Generate Go API client from OpenAPI spec"

generate-client:
	@echo "Generating Go API client..."
	go tool oapi-codegen --config client/sdk/oapi-codegen.yaml doc/openapi.json

install:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing frontend dependencies..."
	cd app && npm install
	@echo "Configuring git hooks..."
	git config core.hooksPath .githooks
	@echo "Git hooks configured to use .githooks directory"

build-frontend:
	@echo "Building frontend..."
	cd app && npm run build
	@echo "Copying frontend build to cmd/microchat-server/static for embedding..."
	rm -rf cmd/microchat-server/static/*
	mkdir -p cmd/microchat-server/static
	cp -r app/dist/* cmd/microchat-server/static/

build-server:
	@echo "Building server..."
	go build -o bin/server ./cmd/microchat-server

build-microchat:
	@echo "Building microchat client..."
	go build -o bin/microchat ./cmd/microchat

build: build-frontend build-server build-microchat
	@echo "Build complete!"

dev:
	@echo "Starting development server..."
	cd app && npm run dev &
	go run ./cmd/microchat-server

run: build
	@echo "Starting server..."
	./bin/server

docker-build:
	@echo "Building Docker image..."
	docker build -t microchat:latest .

docker-up:
	@echo "Starting with docker-compose..."
	docker-compose up -d

docker-down:
	@echo "Stopping docker-compose..."
	docker-compose down

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

lint:
	@echo "Running formatters and linters..."
	go fmt ./...
	go tool modernize -fix ./...
	golangci-lint run --config .golangci.yml
	cd app && npm run format
	cd app && npm run lint

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf app/dist
	rm -rf cmd/microchat-server/static/*
	go clean
