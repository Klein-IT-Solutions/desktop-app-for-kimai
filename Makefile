# Kimai Tracker – developer entry points.
#
#   make dev      Hot-reload development mode (Go + Vite)
#   make build    Production binary in ./bin
#   make test     Go unit tests + frontend type check
#
# Ubuntu 22.04 (no GTK4/WebKitGTK 6.0): append GTK3=1 to use the
# GTK3 / WebKit2GTK 4.1 stack, e.g. `make build GTK3=1`.

APP     := kimai-tracker

# `go install` puts wails3 into $GOBIN (default: $(go env GOPATH)/bin), which is
# often not on PATH. Prepend it so both make and the Wails Taskfiles find it.
GOBIN   := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)
export PATH := $(GOBIN):$(PATH)
WAILS   ?= wails3

ifeq ($(GTK3),1)
TAGS := gtk3
endif

.PHONY: help deps check-wails dev build run test test-go test-integration check-frontend lint package clean kimai-up kimai-down

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

deps: ## Install the Wails CLI and frontend dependencies
	go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.28
	cd frontend && npm ci

check-wails:
	@command -v $(WAILS) >/dev/null 2>&1 || { echo "wails3 not found in $(GOBIN) – run 'make deps' first."; exit 1; }

dev: check-wails ## Run in development mode with hot reload
	EXTRA_TAGS=$(TAGS) $(WAILS) dev -config ./build/config.yml

build: check-wails ## Build a production binary into ./bin
	$(WAILS) build EXTRA_TAGS=$(TAGS)

run: build ## Build and start the app
	./bin/$(APP)

test: test-go check-frontend ## Run all tests

test-go: ## Run Go unit tests
	go test -race ./internal/...

test-integration: ## Run API tests against a real Kimai (KIMAI_TEST_URL, KIMAI_TEST_TOKEN)
	@test -n "$$KIMAI_TEST_URL" -a -n "$$KIMAI_TEST_TOKEN" || { echo "Set KIMAI_TEST_URL and KIMAI_TEST_TOKEN (see 'make kimai-up')."; exit 1; }
	go test -tags integration -run Integration -count=1 -v ./internal/kimai/

kimai-up: ## Start a local Kimai test instance in Docker (seeded, with API token)
	scripts/kimai-test-instance.sh up

kimai-down: ## Stop the Kimai test instance and delete its data
	scripts/kimai-test-instance.sh down

check-frontend: ## Type-check the Svelte/TypeScript frontend
	cd frontend && npm run check

lint: ## go vet + gofmt check
	go vet ./internal/...
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:"; gofmt -l .; exit 1)

package: check-wails ## Build .deb and AppImage into ./bin
	$(WAILS) task linux:create:deb EXTRA_TAGS=$(TAGS)
	$(WAILS) task linux:create:appimage EXTRA_TAGS=$(TAGS)

clean: ## Remove build output
	rm -rf bin frontend/dist .task
