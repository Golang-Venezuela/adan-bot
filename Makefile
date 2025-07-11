GO ?= go

module := $(shell $(GO) list -m)
PROJECT ?= $(notdir $(module))
DOCKER_IMAGE ?= go-ve/adan-bot

goFiles := $(shell find . -iname "*.go" -type f | grep -v "/_" | grep -v "^\./vendor")
goFilesSrc := $(shell $(GO) list -f '{{ range .GoFiles }}{{ $$.Dir }}/{{ . }} {{ end }}' ./...)
goFilesTest := $(shell $(GO) list -f "{{ range .TestGoFiles }}{{ $$.Dir }}/{{ . }} {{ end }}{{ range .XTestGoFiles }}{{ $$.Dir }}/{{ . }} {{ end }}" ./...)

.PHONY: all
all: build

.PHONY: build
build:
	$(GO) build -ldflags="-s -w" -trimpath -o ./dist/ ./...

.PHONY: build-docker
build-docker:
	docker build -t $(DOCKER_IMAGE) .

.PHONY: clean
clean:
	rm -rf dist/

# Development

BENCHMARK_COUNT ?= 1
BENCHMARK_FILE ?= benchmarks-dev.txt
COVERAGE_FILE ?= coverage-dev.txt
CPUPROFILE ?= cpu.prof
ENV_FILE ?= .env
MEMPROFILE ?= mem.prof
TARGET_FUNC ?= .
TARGET_PKG ?= ./...
UID ?= $(shell id -u)
WATCH_TARGET ?= run

benchmarkWebFile := $(shell mktemp -u)-$(PROJECT).html

ifneq "$(notdir $(TARGET_PKG))" "..."
	profileFlags := -cpuprofile "$(CPUPROFILE)" -memprofile "$(MEMPROFILE)"
endif

.PHONY: benchmark
benchmark:
	$(GO) test -v -run none \
		-bench "$(TARGET_FUNC)" -benchmem -count $(BENCHMARK_COUNT) \
		$(profileFlags) \
		"$(TARGET_PKG)" | tee "$(BENCHMARK_FILE)"

.PHONY: benchmark-check
benchmark-check: benchmarks.txt $(BENCHMARK_FILE)
	benchstat "$<" "$(BENCHMARK_FILE)"

.PHONY: benchmark-web
benchmark-web: benchmarks.txt $(BENCHMARK_FILE)
	benchstat -html "$<" "$(BENCHMARK_FILE)" > "$(benchmarkWebFile)"
	xdg-open "$(benchmarkWebFile)"

define benchmarks_file
	BENCHMARK_FILE="$(1)" $(MAKE) -s benchmark
endef

benchmarks.txt:
	$(call benchmarks_file,$@)

benchmarks-%.txt:
	$(call benchmarks_file,$@)

.PHONY: build-docker-debug
build-docker-debug:
	docker build --target debug -t $(DOCKER_IMAGE):debug .

.PHONY: build-docker-dev
build-docker-dev:
	docker build -f dev.Dockerfile -t $(DOCKER_IMAGE):dev .

.PHONY: ca
ca:
	golangci-lint run

.PHONY: ca-fast
ca-fast:
	golangci-lint run --fast

.PHONY: ci
ci: build test lint ca

.PHONY: ci-race
ci-race: build test-race lint ca

.PHONY: clean-dev
clean-dev: clean
	rm -rf "$(CPUPROFILE)" "$(MEMPROFILE)" benchmarks-*.txt coverage-*.txt *.test

.PHONY: coverage
coverage: $(COVERAGE_FILE)
	$(GO) tool cover -func "$(COVERAGE_FILE)"

.PHONY: coverage-check
coverage-check: coverage.txt $(COVERAGE_FILE)
	#coverstat "$<" "$(COVERAGE_FILE)"

.PHONY: coverage-web
coverage-web: $(COVERAGE_FILE)
	$(GO) tool cover -html "$(COVERAGE_FILE)"

define coverage_file
	COVERAGE_FILE="$(1)" $(MAKE) -s test
endef

coverage.txt:
	$(call coverage_file,$@)

coverage-%.txt:
	$(call coverage_file,$@)

GODOC_PORT ?= 6060

.PHONY: doc
doc:
	@echo "Go to http://localhost:$(GODOC_PORT)/pkg/$(module)/"
	godoc -http ":$(GODOC_PORT)" -play
	#GOPROXY=$(shell go env GOPROXY) pkgsite -http ":$(GODOC_PORT)" -cache -proxy

.PHONY: dev-env
dev-env:
	@if ! docker image inspect $(DOCKER_IMAGE):dev > /dev/null 2>&1; then \
		$(MAKE) build-docker-dev; \
	fi
	docker run --rm -it --network host -u "$(UID)" --env-file "$(ENV_FILE)" \
		-v "$$HOME/.cache:/.cache" -v "$$HOME/go/pkg:/go/pkg" -v .:/src \
		$(DOCKER_IMAGE):dev

.PHONY: format
format:
	gofmt -s -w -l $(goFiles)

.PHONY: fuzz
fuzz:
	$(GO) test -v -run none -fuzz "$(TARGET_FUNC)" "$(TARGET_PKG)"

.PHONY: lint
lint:
	gofmt -d -e -s $(goFiles)

.PHONY: run
run:
	go run ./cmd/adan-bot/...

.PHONY: run-race
run-race:
	go run -race ./cmd/adan-bot/...

.PHONY: test
test:
	$(GO) test -v \
		-run "$(TARGET_FUNC)" \
		-coverprofile "$(COVERAGE_FILE)" \
		$(profileFlags) \
		"$(TARGET_PKG)"

.PHONY: test-race
test-race:
	$(GO) test -v -race \
		-run "$(TARGET_FUNC)" \
		-coverprofile "$(COVERAGE_FILE)" \
		$(profileFlags) \
		"$(TARGET_PKG)"


.PHONY: air
air:
	$(GOPATH)/bin/air

.PHONY: help
help:
	@echo "  Available commands:"
	@echo "  all                 \t- Build the project"
	@echo "  build               \t- Build the project"
	@echo "  build-docker        \t- Build Docker image"
	@echo "  build-docker-debug  \t- Build Docker image for debugging"
	@echo "  build-docker-dev    \t- Build Docker image for development"
	@echo "  dev-environment     \t- Run development environment in Docker"
	@echo "  format              \t- Format Go code"
	@echo "  fuzz                \t- Run fuzz tests"
	@echo "  lint                \t- Lint Go code"
	@echo "  run                 \t- Run the application"
	@echo "  run-race            \t- Run the application with race detector"
	@echo "  test                \t- Run tests"
	@echo "  test-race           \t- Run tests with race detector"
	@echo "  air                 \t- Run Air for live reloading"
	@echo "  coverage            \t- Generate code coverage report"
	@echo "  coverage-check      \t- Check code coverage against baseline"
	@echo "  coverage-web        \t- Show code coverage in web browser"
	@echo "  clean               \t- Clean build artifacts"
	@echo "  clean-dev           \t- Clean development artifacts"
	@echo "  doc                 \t- Run Go documentation server"
	@echo "  benchmark           \t- Run benchmarks"
	@echo "  benchmark-check     \t- Check benchmark results"
	@echo "  benchmark-web       \t- Show benchmark results in web browser"
	@echo "  ca                  \t- Run code analysis"
	@echo "  ca-fast             \t- Run code analysis with fast mode"
	@echo "  ci                  \t- Run continuous integration checks"
	@echo "  ci-race             \t- Run continuous integration checks with race detector"
