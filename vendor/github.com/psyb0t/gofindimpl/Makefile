MIN_TEST_COVERAGE := 90
APP_NAME := gofindimpl

.PHONY: all dep lint lint-fix test test-coverage build build-race

all: dep lint-fix test-coverage build ## Run all tasks

dep: ## Get project dependencies
	@echo "Getting project dependencies..."
	@go mod tidy
	@go mod vendor

lint: ## Lint all Golang files
	@echo "Linting all Go files..."
	@go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -test ./...
	@go tool golangci-lint run --timeout=30m0s ./...

lint-fix: ## Lint all Golang files and fix
	@echo "Linting all Go files..."
	@go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix -test ./...
	@go tool golangci-lint run --fix --timeout=30m0s ./...

test: ## Run all tests
	@echo "Running all tests..."
	@go test -race ./...

test-coverage: ## Run tests with coverage check. Fails if coverage is below the threshold.
	@echo "Running tests with coverage check..."
	@trap 'rm -f coverage.txt' EXIT; \
	go test -race -coverprofile=coverage.txt $$(go list ./...); \
	if [ $$? -ne 0 ]; then \
		echo "Test failed. Exiting."; \
		exit 1; \
	fi; \
	result=$$(go tool cover -func=coverage.txt | grep -oP 'total:\s+\(statements\)\s+\K\d+' || echo "0"); \
	if [ $$result -eq 0 ]; then \
		echo "No test coverage information available."; \
		exit 0; \
	elif [ $$result -lt $(MIN_TEST_COVERAGE) ]; then \
		echo "FAIL: Coverage $$result% is less than the minimum $(MIN_TEST_COVERAGE)%"; \
		exit 1; \
	fi

build: ## Build the app binary using Docker (optimized, no debug info)
	@echo "Building $(APP_NAME) binary (optimized)..."
	@mkdir -p ./build
	@HOST_UID=$$(id -u) && HOST_GID=$$(id -g) && \
	echo "Using HOST_UID=$$HOST_UID, HOST_GID=$$HOST_GID" && \
	docker run --rm \
		-v $(PWD):/app \
		-w /app \
		-e HOST_UID=$$HOST_UID \
		-e HOST_GID=$$HOST_GID \
		golang:1.24.6-alpine sh -c "echo 'Starting build...' && \
		CGO_ENABLED=0 go build -v -ldflags='-w -s' -o ./build/$(APP_NAME) ./... && \
		echo 'Build complete, setting ownership...' && \
		chown -R \$$HOST_UID:\$$HOST_GID /app/build && \
		echo 'Done!'"

build-race: ## Build the app binary using Docker (with race detection)
	@echo "Building $(APP_NAME) binary with race detection..."
	@mkdir -p ./build
	@HOST_UID=$$(id -u) && HOST_GID=$$(id -g) && \
	echo "Using HOST_UID=$$HOST_UID, HOST_GID=$$HOST_GID" && \
	docker run --rm \
		-v $(PWD):/app \
		-w /app \
		-e HOST_UID=$$HOST_UID \
		-e HOST_GID=$$HOST_GID \
		golang:1.24.6-alpine sh -c "echo 'Installing build dependencies...' && \
		apk add --no-cache gcc musl-dev && \
		echo 'Starting build with race detection...' && \
		CGO_ENABLED=1 go build -v -race -o ./build/$(APP_NAME) ./... && \
		echo 'Build complete, setting ownership...' && \
		chown -R \$$HOST_UID:\$$HOST_GID /app/build && \
		echo 'Done!'"

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
