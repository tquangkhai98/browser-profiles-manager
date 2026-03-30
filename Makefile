# bpm — Browser Profiles Manager
# Build, test, and develop commands

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  = -ldflags "-s -w -X github.com/tquangkhai98/browser-profiles-manager/cmd.version=$(VERSION) -X github.com/tquangkhai98/browser-profiles-manager/cmd.commit=$(COMMIT) -X github.com/tquangkhai98/browser-profiles-manager/cmd.date=$(DATE)"

.PHONY: build test lint clean install desktop dev

## Build CLI binary with version info
build:
	go build $(LDFLAGS) -o bpm .

## Run all tests
test:
	go test ./... -v -count=1

## Run tests with coverage report
cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Run staticcheck linter
lint:
	@which staticcheck > /dev/null 2>&1 || go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...

## Remove build artifacts
clean:
	rm -f bpm bpm.exe
	rm -f coverage.out coverage.html
	rm -rf build/

## Install to GOPATH/bin
install:
	go install $(LDFLAGS) .

## Build desktop app (Wails)
desktop:
	cd desktop && wails build

## Run desktop app in dev mode
dev:
	cd desktop && wails dev

## Show help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build    Build CLI binary with version info"
	@echo "  test     Run all tests"
	@echo "  cover    Run tests with coverage report"
	@echo "  lint     Run staticcheck linter"
	@echo "  clean    Remove build artifacts"
	@echo "  install  Install to GOPATH/bin"
	@echo "  desktop  Build desktop app (Wails)"
	@echo "  dev      Run desktop app in dev mode"
	@echo "  help     Show this help"
