VENUE_DIR := services/venue-service
VENUE_MIGRATIONS_DIR := $(VENUE_DIR)/db/migrations

.PHONY: install \
	venue-migrate-up \
	venue-migrate-down \
	venue-migrate-status \
	venue-sqlc \
	venue-run \
	venue-test

install:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

venue-migrate-up:
	goose -dir $(VENUE_MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" up

venue-migrate-down:
	goose -dir $(VENUE_MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" down

venue-migrate-status:
	goose -dir $(VENUE_MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" status

venue-sqlc:
	cd $(VENUE_DIR) && sqlc generate

venue-run:
	cd $(VENUE_DIR) && go run cmd/main.go

venue-test:
	cd $(VENUE_DIR) && go test ./...
