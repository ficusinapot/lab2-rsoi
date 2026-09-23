.PHONY: build run unit-test e2e-test test test-all lint lint-fix style-check fmt fmt-check tidy vuln docker-build check

CONFIG ?= config.yaml
GOLANGCI_LINT ?= $(if $(wildcard $(HOME)/go/bin/golangci-lint),$(HOME)/go/bin/golangci-lint,golangci-lint)
GOTESTSUM ?= $(if $(wildcard $(HOME)/go/bin/gotestsum),$(HOME)/go/bin/gotestsum,gotestsum)
GOVULNCHECK ?= $(if $(wildcard $(HOME)/go/bin/govulncheck),$(HOME)/go/bin/govulncheck,govulncheck)

build:
	go build -o bin/gateway -v ./cmd/gateway

run:
	go run ./cmd/gateway --config "$(CONFIG)"

unit-test:
	mkdir -p bin
	GOWORK=off $(GOTESTSUM) --junitfile bin/unit.xml $(TEST_REPORT_ARGS) -- -race -timeout 30s ./...

e2e-test:
	mkdir -p bin
	GATEWAY_PROJECT_ROOT="$(abspath ../..)" GOWORK=off $(GOTESTSUM) --junitfile bin/e2e.xml $(TEST_REPORT_ARGS) -- -tags=e2e -count=1 -timeout 10m ./tests/e2e/...

test: unit-test
test-all: unit-test e2e-test

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
	docker build -f Dockerfile -t rsoi-gateway-service ../..

check: tidy fmt-check lint vuln unit-test build
