.PHONY: help dev build test lint clean docker-up docker-down migrate

# Default target
help:
	@echo "Fiducia - Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  make dev          - Start development environment (Docker)"
	@echo "  make backend      - Run backend locally"
	@echo "  make frontend     - Run frontend locally"
	@echo ""
	@echo "Build:"
	@echo "  make build        - Build all services"
	@echo "  make build-backend - Build Go backend"
	@echo "  make build-frontend - Build Next.js frontend"
	@echo ""
	@echo "Testing:"
	@echo "  make test         - Run all tests"
	@echo "  make test-backend - Run backend tests"
	@echo "  make lint         - Run linters"
	@echo ""
	@echo "Database:"
	@echo "  make migrate      - Run database migrations"
	@echo "  make db-reset     - Reset database (WARNING: deletes data)"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up    - Start Docker services"
	@echo "  make docker-down  - Stop Docker services"
	@echo "  make docker-logs  - View Docker logs"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean        - Remove build artifacts"

# Development
dev:
	docker-compose up -d postgres redis
	@echo "Waiting for services..."
	@sleep 3
	@echo "Starting backend..."
	cd backend && go run ./cmd/server &
	@echo "Starting frontend..."
	cd frontend && npm run dev

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev

# Build
build: build-backend build-frontend

build-backend:
	cd backend && CGO_ENABLED=0 go build -o bin/server ./cmd/server

build-frontend:
	cd frontend && npm run build

# Testing
test: test-backend

test-backend:
	cd backend && go test -v ./...

lint:
	cd backend && go vet ./...
	cd backend && golangci-lint run
	cd frontend && npm run lint

# Database
migrate:
	cd backend && go run ./cmd/server -migrate

db-reset:
	docker-compose exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS fiducia;"
	docker-compose exec postgres psql -U postgres -c "CREATE DATABASE fiducia;"
	@echo "Database reset. Run 'make migrate' to apply migrations."

# Docker
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-build:
	docker-compose build

# Cleanup
clean:
	rm -rf backend/bin
	rm -rf frontend/.next
	rm -rf frontend/node_modules/.cache
