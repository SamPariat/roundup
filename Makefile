VENUE_DIR := services/venue-service
VENUE_MIGRATIONS_DIR := $(VENUE_DIR)/db/migrations
AUTH_DIR := services/auth-service
AUTH_MIGRATIONS_DIR := $(AUTH_DIR)/db/migrations
GATEWAY_DIR := services/gateway
USER_DIR := services/user-service
USER_MIGRATIONS_DIR := $(USER_DIR)/db/migrations
SERVICE_DIRS := $(AUTH_DIR) $(GATEWAY_DIR) $(USER_DIR) $(VENUE_DIR)

# Selects services/<name>/.env.<APP_ENV>. Override per command:
#   make user-run APP_ENV=prod
APP_ENV ?= beta
export APP_ENV

.PHONY: install \
	setup \
	jwt-keys \
	auth-migrate-up \
	auth-migrate-down \
	auth-migrate-status \
	auth-sqlc \
	auth-run \
	auth-test \
	venue-migrate-up \
	venue-migrate-down \
	venue-migrate-status \
	venue-sqlc \
	venue-run \
	venue-test \
	gateway-run \
	gateway-test \
	user-migrate-up \
	user-migrate-down \
	user-migrate-status \
	user-sqlc \
	user-run \
	user-test

install:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Creates each service's .env.beta from its template. Existing files are kept.
# .env.prod is never generated: copy .env.prod.example by hand and fill in real values.
setup:
	@for dir in $(SERVICE_DIRS); do \
		if [ -f $$dir/.env.beta ]; then \
			echo "keep    $$dir/.env.beta"; \
		else \
			cp $$dir/.env.beta.example $$dir/.env.beta && echo "created $$dir/.env.beta"; \
		fi; \
	done

# Generates an Ed25519 keypair for access tokens: the private key goes to
# auth-service (signs), the public key to the gateway (verifies). Run after
# `make setup`. Keeps an existing key; delete JWT_PRIVATE_KEY to regenerate both.
jwt-keys:
	@if grep -q '^JWT_PRIVATE_KEY=".\{1,\}"' $(AUTH_DIR)/.env.beta 2>/dev/null; then \
		echo "keep    JWT keys (delete JWT_PRIVATE_KEY in $(AUTH_DIR)/.env.beta to regenerate)"; \
	else \
		key=$$(openssl genpkey -algorithm ed25519) && \
		pub=$$(printf '%s\n' "$$key" | openssl pkey -pubout) && \
		{ grep -v '^JWT_PRIVATE_KEY=' $(AUTH_DIR)/.env.beta 2>/dev/null; printf 'JWT_PRIVATE_KEY="%s"\n' "$$(printf '%s' "$$key" | awk 'BEGIN{ORS="\\n"} {print}')"; } > $(AUTH_DIR)/.env.beta.tmp && \
		mv $(AUTH_DIR)/.env.beta.tmp $(AUTH_DIR)/.env.beta && \
		{ grep -v '^JWT_PUBLIC_KEY=' $(GATEWAY_DIR)/.env.beta 2>/dev/null; printf 'JWT_PUBLIC_KEY="%s"\n' "$$(printf '%s' "$$pub" | awk 'BEGIN{ORS="\\n"} {print}')"; } > $(GATEWAY_DIR)/.env.beta.tmp && \
		mv $(GATEWAY_DIR)/.env.beta.tmp $(GATEWAY_DIR)/.env.beta && \
		echo "created JWT keypair: private → $(AUTH_DIR)/.env.beta, public → $(GATEWAY_DIR)/.env.beta"; \
	fi

auth-migrate-up:
	goose -dir $(AUTH_MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" up

auth-migrate-down:
	goose -dir $(AUTH_MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" down

auth-migrate-status:
	goose -dir $(AUTH_MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" status

auth-sqlc:
	cd $(AUTH_DIR) && sqlc generate

auth-run:
	cd $(AUTH_DIR) && go run ./cmd

auth-test:
	cd $(AUTH_DIR) && go test ./...

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

gateway-run:
	cd $(GATEWAY_DIR) && go run ./cmd

gateway-test:
	cd $(GATEWAY_DIR) && go test ./...

user-migrate-up:
	goose -dir $(USER_MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" up

user-migrate-down:
	goose -dir $(USER_MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" down

user-migrate-status:
	goose -dir $(USER_MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" status

user-sqlc:
	cd $(USER_DIR) && sqlc generate

user-run:
	cd $(USER_DIR) && go run ./cmd

user-test:
	cd $(USER_DIR) && go test ./...
