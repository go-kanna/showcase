DB_USER = admin
DB_NAME = showcase
DB_URL = postgres://admin:password@localhost:5432/showcase?sslmode=disable
GOOSE_MIGRATION_DIR=internal/infra/storage/db/migrate/sql

SQL_PATHS = internal/infra/storage/db/migrate/sql

.PHONY: test
test:
	go test ./... -race -coverprofile=coverage.out -covermode=atomic

.PHONY: lint
lint:
	golangci-lint run

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix

.PHONY: sql-lint
sql-lint:
	sqlfluff lint --disable-progress-bar $(SQL_PATHS)

.PHONY: sql-fix
sql-fix:
	sqlfluff fix --disable-progress-bar $(SQL_PATHS)
	$(MAKE) gen-sqlc

.PHONY: gen
gen: gen-go gen-proto gen-sqlc

.PHONY: gen-go
gen-go:
	go generate ./...

.PHONY: gen-check
gen-check:
	KANNA_CHECK=1 go generate ./...

.PHONY: gen-proto
gen-proto:
	cd proto && buf generate

.PHONY: gen-sqlc
gen-sqlc:
	go tool sqlc generate

.PHONY: db-create
db-create:
	docker compose exec -T db psql -U $(DB_USER) -d postgres -tAc \
		"SELECT 1 FROM pg_database WHERE datname = '$(DB_NAME)'" | grep -q 1 || \
		docker compose exec -T db createdb -U $(DB_USER) $(DB_NAME)

.PHONY: db-migrate
db-migrate:
	GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) go tool goose postgres $(DB_URL) up

.PHONY: db-drop
db-drop:
	docker compose exec -T db dropdb -U $(DB_USER) --force --if-exists $(DB_NAME)

.PHONY: db-reset
db-reset: db-drop db-create db-migrate

.POHNY: new-migration
new-migration:
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make new-migration name=migration_name"; \
		exit 1; \
	fi
	go tool goose -dir internal/infra/storage/db/migrate/sql create $(name) sql
