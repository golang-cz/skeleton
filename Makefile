.PHONY: all run test build build-all vendor

SHELL=bash -o pipefail -e
GOBIN ?= $$PWD/bin

export GOFLAGS=-mod=vendor
export CGO_ENABLED=0
export GOGC=off
export GOPRIVATE=github.com/vcilabs

all:
	@echo "  run-api          - run Skeleton API in dev mode"
	@echo "  dev              - start postgresql and run Skeleton API"
	@echo "  dev-jq           - start postgresql and run Skeleton API with JQ pipe"
	@echo "  test             - run tests"
	@echo "  build            - build all binaries on current OS/Arch"
	@echo "  docker-build     - build Docker image"
	@echo "  vendor           - vendor third party Go modules"
	@echo "  create-migration - create sql/go migration"
	@echo "  db-up            - update state to last migration"
	@echo "  db-down          - downgrade state of migration"
	@echo "  init             - install tools, vendor Go modules and build packages"
	@echo "  start            - run docker-compose up and run migrations"

init: tools vendor build
start: up db-up
stop: down
clean: db-reset down

generate:
	go generate -x ./...

tools:
	GOFLAGS="" go install github.com/VojtechVitek/rerun/cmd/rerun@latest

vendor:
	go mod tidy && go mod vendor && go mod tidy

build:
	$(call build, ./cmd/api ./cmd/goose)

build-goose:
	$(call build, ./cmd/goose)

dev-db-up:
	@docker-compose -f docker-compose.yaml up -d postgresql

dev-db-down:
	@docker-compose -f docker-compose.yaml down 

run-api:
	rerun -watch ./ -ignore vendor bin tests mounts -run sh -c 'go build -o ./bin/api ./cmd/api/main.go && ./bin/api -config=etc/config.toml'

run-scheduler:
	rerun -watch ./ -ignore vendor bin tests mounts -run sh -c 'go build -o ./bin/scheduler ./cmd/scheduler/main.go && ./bin/scheduler -config=etc/config.toml'

dev: dev-db-up db-up
	rerun -watch ./ -ignore vendor bin tests mounts -run sh -c 'go build -o ./bin/api ./cmd/api/main.go && ./bin/api -config=etc/config.toml'

dev-jq: dev-db-up db-up
	rerun -watch ./ -ignore vendor bin tests mounts -run sh -c 'go build -o ./bin/api ./cmd/api/main.go && ./bin/api -config=etc/config.toml | jq -S'

# DOCKER

up:
	@docker-compose up -d --remove-orphans

up-build:
	@docker-compose up -d --remove-orphans --build

down:
	@docker-compose down --remove-orphans

down-rmi:
	@docker-compose down --remove-orphans --rmi all


test-analysis:
	docker run --rm -v $(shell pwd):/app -u $(shell id -u):$(shell id -g) -w /app -e GOCACHE=/app/.cache/golang -e GOLANGCI_LINT_CACHE=/app/.cache/golangci ghcr.io/golang-cz/static-analysis:latest golangci-lint run  -c .golangci.yml services/... pkg/... 

# GOOSE DB Migrations
create-migration: build-goose
	@./bin/goose -config=./etc/config.toml create $(filter-out $@,$(MAKECMDGOALS))

db-update-schema:
	@./scripts/pg_dump.sh convo --schema-only | grep -v -e '^--' -e '^COMMENT ON' -e '^REVOKE' -e '^GRANT' -e '^SET' -e 'ALTER DEFAULT PRIVILEGES' -e 'OWNER TO' | cat -s > ./data/schema/schema.sql

db-up: build-goose
	@./bin/goose -config=./etc/config.toml up

db-down:
	@./bin/goose -config=./etc/config.toml down

db-down-to: 
	@./bin/goose -config=./etc/config.toml down-to $(MIGRATION_VERSION)

db-reset: 
	@./bin/goose -config=./etc/config.toml reset

db-status:
	@./bin/goose -config=./etc/config.toml status

define build
	mkdir -p $(GOBIN)
	GOGC=off GOBIN=$(GOBIN) \
	     go install -v \
	     -tags='$(BUILDTAGS)' \
	     -gcflags='-e' \
	     $(1)
endef

define run
	rerun -watch ./ -ignore vendor bin tests data/schema *.sqlc $$(ls -d data/emails/templates/* ) $$(ls -d cmd/* | grep -v $(1)) -run sh -c 'GOGC=off go build -o ./bin/$(1) ./cmd/$(1)/main.go && ./bin/$(1) -config=etc/dev.toml'
endef
