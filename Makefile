.PHONY: help dev db-up db-down db-reset backend seed-dev test test-v lint tidy

# Default target
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  dev        Start Postgres and run the backend server"
	@echo "  db-up      Start Postgres (detached)"
	@echo "  db-down    Stop Postgres"
	@echo "  db-reset   Wipe the Postgres volume and restart fresh"
	@echo "  backend    Run the backend server (DB must already be up)"
	@echo "  seed-dev   Seed ~75 problem attempts for the dev user (idempotent)"
	@echo "  test       Run all Go tests"
	@echo "  test-v     Run all Go tests with verbose output"
	@echo "  lint       Run go vet on all packages"
	@echo "  tidy       Tidy go.mod and go.sum"

# Start DB then run the server (Ctrl-C kills both)
dev: db-up
	$(MAKE) backend

db-up:
	docker compose up -d
	@echo "Waiting for Postgres to be ready..."
	@until docker compose exec -T postgres pg_isready -U interviewiq -d interviewiq > /dev/null 2>&1; do sleep 1; done
	@echo "Postgres is ready."

db-down:
	docker compose down

# Tear down containers AND volumes — destroys all data
db-reset:
	docker compose down -v
	docker compose up -d
	@echo "Waiting for Postgres to be ready..."
	@until docker compose exec -T postgres pg_isready -U interviewiq -d interviewiq > /dev/null 2>&1; do sleep 1; done
	@echo "Fresh database is ready."

backend:
	cd backend && go run ./cmd/server

# Seed ~75 realistic problem attempts for clerk_user_id = "dev_seed_user".
# Idempotent: exits early if the user already has problems.
seed-dev: db-up
	cd backend && go run ./cmd/seed-dev

test:
	cd backend && go test ./...

test-v:
	cd backend && go test -v ./...

lint:
	cd backend && go vet ./...

tidy:
	cd backend && go mod tidy
