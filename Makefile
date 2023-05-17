.PHONY: all run test build build-all vendor

SHELL=bash -o pipefail -e

export GOFLAGS=-mod=vendor
export CGO_ENABLED=0
export GOGC=off
export GOPRIVATE=github.com/vcilabs

all:
	@echo "  run-api          - run Convo API in dev mode"
	@echo "  test             - run tests"
	@echo "  build            - build all binaries on current OS/Arch"
	@echo "  docker-build     - build Docker image"
	@echo "  vendor           - vendor third party Go modules"
	@echo "  create-migration - create sql/go migration"
	@echo "  db-up - update state to last migration"
	@echo "  db-down - downgrade state of migration"

# GOOSE

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

db-status:
	@./bin/goose -config=./etc/config.toml status

GOBIN ?= $$PWD/bin

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

##
## Tools
##
init: tools vendor build
start: up db-up

tools:
	# cd outside of this project directory, otherwise `go get' would start updating go.mod file :/
	cd /tmp && GOFLAGS="" go install github.com/VojtechVitek/rerun/cmd/rerun@latest

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

run-api-dev: dev-db-up
	rerun -watch ./ -ignore vendor bin tests mounts -run sh -c 'go build -o ./bin/api ./cmd/api/main.go && ./bin/api -config=etc/config.toml | jq -S'

# DOCKER

up:
	#	@docker network create -d bridge --subnet 172.24.0.0/16 convo &> /dev/null || :
	@docker-compose up -d --remove-orphans

down:
	@docker-compose down --remove-orphans
