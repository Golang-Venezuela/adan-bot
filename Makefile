DATE := $(shell date +%Y)

help: ## Display this help message, listing all available targets and their descriptions.
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1;34m${DOCKER_NAMESPACE}\033[0m\tGolangVzla - Adan Bot\n \n\033[1;32mUsage:\033[0m\n  make \033[1;34m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[1;34m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1;33m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


## Project Metadata & Defaults
GO ?= go
module := $(shell $(GO) list -m)
PROJECT ?= $(notdir $(module))
DOCKER_IMAGE ?= go-ve/adan-bot
UID ?= $(shell id -u)
GODOC_PORT ?= 6060
ENV_FILE ?= .env

CPUPROFILE ?= cpu.prof
MEMPROFILE ?= mem.prof
WATCH_TARGET ?= run

## File Lists
goFiles := $(shell find . -iname "*.go" -type f | grep -v "/_" | grep -v "^\./vendor")
goFilesSrc := $(shell $(GO) list -f '{{ range .GoFiles }}{{ $$.Dir }}/{{ . }} {{ end }}' ./...)
goFilesTest := $(shell $(GO) list -f "{{ range .TestGoFiles }}{{ $$.Dir }}/{{ . }} {{ end }}{{ range .XTestGoFiles }}{{ $$.Dir }}/{{ . }} {{ end }}" ./...)

##@ Build & Clean
.PHONY: all build clean clean-dev
all: build ## Build the project

build: ## Build the project
	cd cmd/adan-bot && $(GO) build -ldflags="-s -w" -trimpath -o ./dist/ ./...

clean: ## Clean build artifacts
	rm -rf dist/

clean-dev: clean ## Clean development artifacts
	rm -rf "$(CPUPROFILE)" "$(MEMPROFILE)" benchmarks-*.txt coverage-*.txt *.test

##@ Execution
.PHONY: run run-race air

run: ## Run the application
	go run ./cmd/adan-bot/...

run-race: ## Run the application with race detector
	go run -race ./cmd/adan-bot/...

air: ## Run Air for live reloading
	$(GOPATH)/bin/air

##@ Docker
.PHONY: build-docker build-docker-debug build-docker-dev dev-env

.PHONY: build-docker
build-docker: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

build-docker-debug: ## Build Docker image for debugging
	docker build --target debug -t $(DOCKER_IMAGE):debug .

build-docker-dev: ## Build Docker image for development
	docker build -f dev.Dockerfile -t $(DOCKER_IMAGE):dev .

dev-env: ## Run development environment in Docker
	@if ! docker image inspect $(DOCKER_IMAGE):dev > /dev/null 2>&1; then \
		$(MAKE) build-docker-dev; \
	fi
	docker run --rm -it --network host -u "$(UID)" --env-file "$(ENV_FILE)" \
		-v "$$HOME/.cache:/.cache" -v "$$HOME/go/pkg:/go/pkg" -v .:/src \
		$(DOCKER_IMAGE):dev

##@ Testing
.PHONY: test test-race
COVERAGE_FILE ?= reports/coverage-dev.txt
TARGET_FUNC ?= .
TARGET_PKG ?= ./...

profileFlags := # dynamically set below
ifneq "$(notdir $(TARGET_PKG))" "..."
	profileFlags := -cpuprofile "$(CPUPROFILE)" -memprofile "$(MEMPROFILE)"
endif

test: ## Run tests
	@$(GO) test -v -run "$(TARGET_FUNC)" -coverprofile "$(COVERAGE_FILE)" $(profileFlags) "$(TARGET_PKG)"

test-race: ## Run tests with race detector
	@$(GO) test -v -race -run "$(TARGET_FUNC)" -coverprofile "$(COVERAGE_FILE)" $(profileFlags) "$(TARGET_PKG)"

##@ Benchmarking
.PHONY: benchmark benchmark-check benchmark-web
BENCHMARK_COUNT ?= 1
BENCHMARK_FILE ?= benchmarks-dev.txt
benchmarkWebFile := $(shell mktemp -u)-$(PROJECT).html

benchmark: ## Run benchmarks
	$(GO) test -v -run none -bench "$(TARGET_FUNC)" -benchmem -count $(BENCHMARK_COUNT) $(profileFlags) "$(TARGET_PKG)" | tee "$(BENCHMARK_FILE)"

benchmark-check: benchmarks.txt $(BENCHMARK_FILE) ## Check benchmark results
	benchstat "$<" "$(BENCHMARK_FILE)"

benchmark-web: benchmarks.txt $(BENCHMARK_FILE) ## Show benchmark results in web browser
	benchstat -html "$<" "$(BENCHMARK_FILE)" > "$(benchmarkWebFile)"
	xdg-open "$(benchmarkWebFile)"

define benchmarks_file
	BENCHMARK_FILE="$(1)" $(MAKE) -s benchmark
endef

benchmarks.txt:
	$(call benchmarks_file,$@)

benchmarks-%.txt:
	$(call benchmarks_file,$@)

##@ Coverage
.PHONY: coverage coverage-check coverage-web

coverage: $(COVERAGE_FILE) ## Generate code coverage report
	$(GO) tool cover -func "$(COVERAGE_FILE)"

coverage-check: coverage.txt $(COVERAGE_FILE) ## Check code coverage against baseline
	#coverstat "$<" "$(COVERAGE_FILE)"

coverage-web: $(COVERAGE_FILE) ## Show code coverage in web browser
	$(GO) tool cover -html "$(COVERAGE_FILE)"

define coverage_file
	COVERAGE_FILE="$(1)" $(MAKE) -s test
endef

coverage.txt:
	$(call coverage_file,$@)

coverage-%.txt:
	$(call coverage_file,$@)

##@ Formatting & Linting
.PHONY: format lint ca ca-fast

format: ## Format Go code
	gofmt -s -w -l $(goFiles)

lint: ## Lint Go code
	gofmt -d -e -s $(goFiles)

ca: ## Run code analysis
	golangci-lint --config .golangci.yml run

ca-fast: ## Run code analysis with fast mode
	golangci-lint --config .golangci.yml run --fast

##@ Fuzzing
.PHONY: fuzz

fuzz: ## Run fuzz tests
	$(GO) test -v -run none -fuzz "$(TARGET_FUNC)" "$(TARGET_PKG)"

##@ CI
.PHONY: ci ci-race
ci: build test lint ca ## Run continuous integration checks

ci-race: build test-race lint ca ## Run continuous integration checks with race detector

##@ Documentation
.PHONY: doc
doc: ## Run Go documentation server
	@echo "Stopping any running pkgsite processes..."
	@pkill pkgsite || true
	@echo "Cleaning Go module cache..."
	@go clean -modcache && rm -rf $$GOPATH/pkg
	@echo "Tidying up Go modules..."
	@go mod tidy
	@echo "Launching pkgsite at http://localhost:8080 ..."
	@nohup pkgsite > /dev/null 2>&1 &
	@sleep 2
	@open -a "Google Chrome" http://localhost:8080/$(module)
