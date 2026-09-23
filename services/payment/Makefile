.PHONY: generate build run migrate migrate-up migrate-status migrate-hash migrations-check unit-test integration-test test test-all lint lint-fix style-check fmt fmt-check tidy vuln docker-build check

CONFIG ?= config.yaml
TEST_CONFIG ?= config.test.yaml
MIGRATION_CONFIG ?= migrations.yaml
GOLANGCI_LINT ?= $(if $(wildcard $(HOME)/go/bin/golangci-lint),$(HOME)/go/bin/golangci-lint,golangci-lint)
GOTESTSUM ?= $(if $(wildcard $(HOME)/go/bin/gotestsum),$(HOME)/go/bin/gotestsum,gotestsum)
GOVULNCHECK ?= $(if $(wildcard $(HOME)/go/bin/govulncheck),$(HOME)/go/bin/govulncheck,govulncheck)
ATLAS ?= atlas
YQ ?= $(if $(wildcard $(HOME)/go/bin/yq),$(HOME)/go/bin/yq,yq)

generate:
	GOWORK=off go generate -v ./ent

build:
	go build -o bin/payment -v ./cmd/payment

run:
	go run ./cmd/payment --config "$(CONFIG)"

migrate migrate-up:
	@set -eu; dsn="$$( $(YQ) -er '.database.url' "$(CONFIG)" )"; dir="$$( $(YQ) -er '.directory' "$(MIGRATION_CONFIG)" )"; lock="$$( $(YQ) -er '.lock_timeout' "$(MIGRATION_CONFIG)" )"; duration="$$( $(YQ) -er '.timeout' "$(MIGRATION_CONFIG)" )"; timeout "$$duration" $(ATLAS) migrate apply --dir "file://$$dir" --url "$$dsn" --lock-timeout "$$lock"

migrate-status:
	@set -eu; dsn="$$( $(YQ) -er '.database.url' "$(CONFIG)" )"; dir="$$( $(YQ) -er '.directory' "$(MIGRATION_CONFIG)" )"; duration="$$( $(YQ) -er '.timeout' "$(MIGRATION_CONFIG)" )"; timeout "$$duration" $(ATLAS) migrate status --dir "file://$$dir" --url "$$dsn"

migrate-hash:
	@set -eu; dir="$$( $(YQ) -er '.directory' "$(MIGRATION_CONFIG)" )"; $(ATLAS) migrate hash --dir "file://$$dir"

migrations-check:
	@set -eu; dir="$$( $(YQ) -er '.directory' "$(MIGRATION_CONFIG)" )"; $(ATLAS) migrate validate --dir "file://$$dir"

unit-test:
	mkdir -p bin
	GOWORK=off $(GOTESTSUM) --junitfile bin/unit.xml $(TEST_REPORT_ARGS) -- -race -timeout 30s ./...

integration-test:
	mkdir -p bin
	GOWORK=off $(GOTESTSUM) --junitfile bin/integration.xml $(TEST_REPORT_ARGS) -- -tags=integration -count=1 -race -timeout 3m ./tests/integration/...

test: unit-test
test-all: unit-test integration-test

lint:
	$(GOLANGCI_LINT) run -c .golangci.yml ./...

lint-fix:
	$(GOLANGCI_LINT) run --fix -c .golangci.yml ./...

style-check:
	$(GOLANGCI_LINT) run --enable-only lll -c .golangci.yml ./...

fmt:
	$(GOLANGCI_LINT) fmt -c .golangci.yml ./...

fmt-check:
	$(GOLANGCI_LINT) fmt --diff -c .golangci.yml ./...

tidy:
	GOWORK=off go mod tidy --diff

vuln:
	$(GOVULNCHECK) -show verbose ./...

docker-build:
	docker build -f Dockerfile -t rsoi-payment-service ../..

check: tidy fmt-check lint vuln unit-test migrations-check
