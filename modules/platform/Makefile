.PHONY: build unit-test test lint lint-fix style-check fmt fmt-check tidy vuln check
GOLANGCI_LINT ?= $(if $(wildcard $(HOME)/go/bin/golangci-lint),$(HOME)/go/bin/golangci-lint,golangci-lint)
GOTESTSUM ?= $(if $(wildcard $(HOME)/go/bin/gotestsum),$(HOME)/go/bin/gotestsum,gotestsum)
GOVULNCHECK ?= $(if $(wildcard $(HOME)/go/bin/govulncheck),$(HOME)/go/bin/govulncheck,govulncheck)

build:
	go build ./...

unit-test:
	$(GOTESTSUM) -- -race -timeout 30s ./...

test: unit-test

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

check: tidy fmt-check lint vuln unit-test build
