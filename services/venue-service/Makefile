MIGRATIONS_DIR := db/migrations

.PHONY: install migrate-up migrate-down migrate-status sqlc run test

install:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" status

sqlc:
	sqlc generate

run:
	go run cmd/main.go

test:
	go test ./...
