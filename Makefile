ENV_FILE ?= .env
ifneq ($(wildcard $(ENV_FILE)),)
include $(ENV_FILE)
export
endif

LOCAL_BIN:=$(CURDIR)/bin
PATH:=$(LOCAL_BIN):$(PATH)

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.PHONY: help

compose-up: ### Run docker-compose
	docker compose up --build -d
.PHONY: compose-up

compose-down: ### Down docker-compose
	docker compose down
.PHONY: compose-down

swag-v1: ### swag init
	swag init -g internal/controller/http/router.go
.PHONY: swag-v1

run: swag-v1 ### Run the application without implicit database migrations
	DISABLE_SWAGGER_HTTP_HANDLER='' GIN_MODE=debug CGO_ENABLED=0 go run ./cmd/app
.PHONY: run

run-with-migrations: swag-v1 ### Apply migrations, then run the application
	CGO_ENABLED=0 go run ./cmd/migrate && \
	DISABLE_SWAGGER_HTTP_HANDLER='' GIN_MODE=debug CGO_ENABLED=0 go run ./cmd/app
.PHONY: run-with-migrations

linter-golangci: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci

linter-hadolint: ### check by hadolint linter
	hadolint Dockerfile Dockerfile.migrate integration-test/Dockerfile
.PHONY: linter-hadolint

linter-dotenv: ### validate the dotenv example
	python3 scripts/validate_env.py .env.example
.PHONY: linter-dotenv

test: ### run test
	go test -v -cover -race ./...
.PHONY: test

migrate-create:  ### create new migration
	migrate create -ext sql -dir migrations 'migrate_name'
.PHONY: migrate-create

migrate-up: ### migration up
	@test -n "$(MYSQL_URL)" || (echo "MYSQL_URL is required; copy .env.example to .env or export it"; exit 1)
	migrate -path migrations -database '$(if $(filter mysql://%,$(MYSQL_URL)),$(MYSQL_URL),mysql://$(MYSQL_URL))' up
.PHONY: migrate-up

bin-deps:
	GOBIN=$(LOCAL_BIN) go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
	GOBIN=$(LOCAL_BIN) go install github.com/swaggo/swag/cmd/swag@v1.16.6

restart-logstash-1: ### Run restart-logstash-1
	docker compose up -d logstash_0
.PHONY: restart-logstash-1

restart-logstash-2: ### Run restart-logstash-2
	docker compose up -d logstash_1
.PHONY: restart-logstash-2
