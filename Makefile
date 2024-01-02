.PHONY: all run test build build-all vendor config

SHELL=bash -o pipefail -e
GOBIN ?= $$PWD/bin

export GOFLAGS=-mod=vendor
export CGO_ENABLED=0
export GOPRIVATE=github.com/golang-cz

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

##
## DEVELOPMENT
##

init: config git-hooks tools vendor build

config:
	cp etc/config.sample.toml etc/config.toml

config-to-sample:
	cp etc/config.toml etc/config.sample.toml

git-hooks:
	ln -s -i ../../scripts/prepare-commit-msg .git/hooks/prepare-commit-msg || true
	ln -s -i ../../scripts/pre-commit .git/hooks/pre-commit || true
	ln -s -i ../../scripts/pre-push .git/hooks/pre-push || true

format:
	goimports -w -local github.com/golang-cz/skeleton $$(find . -type f -name '*.go' -not -name '*.gen.go' -not -name '*.pb.go' -not -name '*.twirp.go' -not -path "./vendor/*")
	gofmt -s -w $$(find . -type f -name '*.go' -not -name '*.gen.go' -not -name '*.pb.go' -not -name '*.twirp.go' -not -path "./vendor/*")

tools:
	GOFLAGS="" go install -v github.com/mfridman/tparse@latest
	GOFLAGS="" go install -v github.com/webrpc/webrpc/cmd/webrpc-gen@latest
	GOFLAGS="" go install golang.org/x/tools/cmd/goimports@latest
	GOFLAGS="" go install github.com/cosmtrek/air@latest

goimports:
	@goimports -l -w -local github.com/golang-cz/skeleton $$(find . -type f -name '*.go' -not -name '*.gen.go' -not -name '*.pb.go' -not -name '*.twirp.go' -not -path "./vendor/*")

vendor:
	go mod tidy && go mod vendor && go mod tidy

todo:
	@git grep TODO -- './*' ':!./vendor/' ':!./Makefile' || :

rebase-generate:
	${MAKE} generate
	git add proto/
	git commit
	git rebase --continue

up:
	docker compose -f docker-compose.yaml up --detach

down:
	docker compose -f docker-compose.yaml down

##
## BUILDS
##

GOBIN ?= $$PWD/bin

define build
	mkdir -p $(GOBIN)
	GOGC=off GOBIN=$(GOBIN) \
	     go install -v \
	     -tags='$(BUILDTAGS)' \
	     -gcflags='-e' \
	     -ldflags='-s -w -X "github.com/golang-cz/skeleton/pkg/version.VERSION=$(VERSION)" -X "github.com/golang-cz/skeleton/pkg/version.VERSION_SHA=$(COMMIT)"' \
	     $(1)
endef

build: build-api build-goose build-scheduler

build-api:
	@$(call build, ./cmd/api)

build-goose:
	@$(call build, ./cmd/goose)

build-scheduler:
	@$(call build, ./cmd/scheduler)

dist: BUILDTAGS += production
dist: build-all

run:
	@awk -F'[ :]' '!/^all:/ && /^run-([A-z_-]+):/ {print "make " $$1}' Makefile

run-api:
	air -c .air.toml --build.cmd "go build -o bin/api cmd/api/main.go" --build.bin "./bin/api"

run-scheduler:
	air -c .air.toml --build.cmd "go build -o bin/scheduler cmd/scheduler/main.go" --build.bin "./bin/scheduler $(filter-out $@,$(MAKECMDGOALS))"

test-analysis:
	docker run --rm -v $(shell pwd):/app -u $(shell id -u):$(shell id -g) -w /app -e GOCACHE=/app/.cache/golang -e GOLANGCI_LINT_CACHE=/app/.cache/golangci ghcr.io/golang-cz/static-analysis:latest golangci-lint run  -c .golangci.yml  pkg/... 

##
## TEST
##

test: test-config-toml test-analysis test-e2e
test-pre-push: test-config-toml test-e2e

test-analysis: test-analysis-basic test-analysis-tagliatelle

test-analysis-basic:
	@echo "Test Analysis"
	@docker run --rm --name skeleton-static-analysis -v $(shell pwd):/app -u $(shell id -u):$(shell id -g) -w /app -e GOCACHE=/app/.cache/golang -e GOLANGCI_LINT_CACHE=/app/.cache/golangci ghcr.io/golang-cz/static-analysis:v23.8.27-17c0922 golangci-lint run -c .golangci.yml

test-analysis-tagliatelle:
	@echo "Test Analysis Tagliatelle"
	@docker run --rm --name skeleton-static-analysis-tagliatelle -v $(shell pwd):/app -u $(shell id -u):$(shell id -g) -w /app -e GOCACHE=/app/.cache/golang -e GOLANGCI_LINT_CACHE=/app/.cache/golangci ghcr.io/golang-cz/static-analysis:v23.8.27-17c0922 golangci-lint run -c .golangci-tagliatelle.yml ./data/... ./proto/...

test-e2e:
	@echo "Test E2E"
	@go test -parallel 1 ./... -json | tparse -all

test-e2e-coverage:
	@echo "Test E2E coverage"
	@go test -cover -coverprofile=tests/coverage.out -parallel 1 ./... -json | tparse -all

test-e2e-coverage-inspect: test-e2e-coverage
	@echo "Test E2E coverage with inspect"
	@go tool cover -html=tests/coverage.out

test-config-toml:
	@echo "Test Config TOMLs"
	@go run scripts/toml_keys_compare/toml_keys_compare.go etc/config.toml etc/config.sample.toml || exit 1

##
## DATABASE
##
create-migration-sql: build-goose
	@./bin/goose -config=./etc/config.toml create $(filter-out $@,$(MAKECMDGOALS)) sql

create-migration-go: build-goose
	@./bin/goose -config=./etc/config.toml create $(filter-out $@,$(MAKECMDGOALS)) go 

db-update-schema:
	docker exec -it skeleton-postgres pg_dump -U devbox -d skeleton --schema-only | grep -v -e '^--' -e '^COMMENT ON' -e '^REVOKE' -e '^GRANT' -e '^SET' -e 'ALTER DEFAULT PRIVILEGES' -e 'OWNER TO' | cat -s > ./db/schema.sql

db-up: build-goose
	@./bin/goose -config=./etc/config.toml up
	@$(MAKE) db-update-schema
	@$(MAKE) db-generate-svg-schema

db-down: build-goose
	@./bin/goose -config=./etc/config.toml down
	@$(MAKE) db-update-schema

db-down-to: build-goose
	e./bin/goose -config=./etc/config.toml down-to $(MIGRATION_VERSION)
	@$(MAKE) db-update-

db-redo: build-goose
	@./bin/goose -config=./etc/config.toml redo

db-status: build-goose
	@./bin/goose -config=./etc/config.toml status

db-version: build-goose
	@./bin/goose -config=./etc/config.toml version

db-reset:
	docker exec -it skeleton-postgres /bin/sh -c "/home/db.sh drop skeleton"
	@$(MAKE) db-create
	@$(MAKE) db-up

db-create:
	docker exec -it skeleton-postgres /bin/sh -c "/home/db.sh create skeleton"


db-generate-svg-schema:
	@docker run -it --rm --name skeleton-db-generate-svg-schema -v $(shell pwd):/app -u $(shell id -u):$(shell id -g) -w /app ghcr.io/golang-cz/sql2diagram:latest sql2diagram -schema /app/db/schema.sql > db/schema.svg

generate:
	go generate -x ./...
	# sed -i '/^type .* struct {$$/ s/\.//g' ./proto/clients/skeleton/skeletonClient.gen.go
	
docs:
	@echo make docs-users
	@echo make godoc

docs-users:
	@echo http://localhost:8088
	docker run --rm --name skeleton-docs-users -p 8088:8080 -v $$PWD/proto/docs:/app -e SWAGGER_JSON=/app/skeletonUsersApi.gen.yaml swaggerapi/swagger-ui

godoc:
	@which pkgsite || go install golang.org/x/pkgsite/cmd/pkgsite@latest
	pkgsite -http=localhost:9933 -open ./

datatype:
	@go run ./scripts/generators/datatype/datatype.go $(filter-out $@,$(MAKECMDGOALS))
