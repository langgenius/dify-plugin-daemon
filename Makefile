.PHONY: test lint test-integration

APP_NAME:=dify-plugin-daemon
GOCMD:=go
GOTEST:=$(GOCMD) test
GOLANGCI_LINT:=golangci-lint
GOLANGCI_CONFIG:=.golangci.yml

## test: Run unit tests (skips integration tests)
test:
	@echo "Running unit tests..."
	$(GOTEST) -v -timeout 10m $(shell go list ./... | grep -v /integration/)

## test-integration: Run integration tests (requires docker-compose)
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -timeout 10m ./integration/...

## test-all: Run all tests (unit + integration)
test-all: test test-integration

## lint: Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	$(GOLANGCI_LINT) run --config $(GOLANGCI_CONFIG) --timeout 5m ./...
