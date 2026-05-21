.PHONY: dev dev-db up down build

# Start only the database (for local dev with go run)
dev-db:
	docker compose up -d db

# Run app locally, connected to db in Docker (run from project root)
dev: dev-db
	go run ./cmd

# Full stack in Docker
up:
	docker-compose up -d --build

down:
	docker compose down -v

build:
	go build -o bin/main ./cmd
