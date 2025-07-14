# Contributing Guide

Any contribution to this project means implicitly that you accept the
[code of conduct](CODE_OF_CONDUCT.md) from this project.

See [Build system](#build-system) for a reference of the build system.

## Requirements

- [Git](https://git-scm.com/)
- [Go](https://golang.org/dl/) >= 1.24

### Optional

- [GNU Make](https://www.gnu.org/software/make/) >= 4.3 (build tool)
- [GolangCI Lint](https://github.com/golangci/golangci-lint/releases) >= 1.54
- [air](https://github.com/cosmtrek/air/) >= 1.49 (☁️ Live reload for Go apps)
- [reflex](https://github.com/cespare/reflex)

## Guidelines

- **Git commit messages:** <https://chris.beams.io/posts/git-commit/>;
  additionally any commit must be scoped to the package where changes were
  made, which is prefixing the message with the package name, e.g.
  `build: Do something`.

- **Git branching model:** <https://guides.github.com/introduction/flow/>.

- **Version number bumping:** <https://semver.org/>.

- **Changelog format:** <http://keepachangelog.com/>.

- **Go code guidelines:** <https://golang.org/doc/effective_go.html>.

## Instructions

1. Create a new branch with a short name that describes the changes that you
   intend to do. If you don't have permissions to create branches, fork the
   project and do the same in your forked copy.

2. Do any change you need to do and add the respective tests.

3. **(Optional)** Run `make ci-race` (or `make ci` if your platform doesn't
   support the Go's race conditions detector) in the project root folder to
   verify that everything is working.

4. Create a [pull request][] to the `main` branch.

[pull request]: https://github.com/Golang-Venezuela/adan-bot/compare

## Build system

The build system provides a set of utilities for improving the developer
experience. Most of them may be typed directly in the terminal and only serve
as example since the Go toolchain provides the majority of tools needed with
a simple interface.

For a complete list of build targets see [Makefile](Makefile) or run `make help`.

**Usage:**

```shell
make [VARIABLE=VALUE...] [TARGET...]
```

### Building

```shell
make build
```

**Variables:**

- `GO`: Go toolchain to use. (default: `go`)

#### Docker

A docker image is also provided.

```shell
make build-docker
```

```shell
docker run --rm -it go-ve/adan-bot
```

The resulting image uses `scratch` as base image, considerably reducing the
image size and improving security by reducing the attack surface, but there are
some cases where having a shell and common commands helps during debugging
process. For this cases, you may use the `build-docker-debug` target.

```shell
make build-docker-debug
```

```shell
docker run --rm -it go-ve/adan-bot:debug
docker run --rm -it go-ve/adan-bot:debug sh  # Launch a shell session.
```

It is also possible to prepare a development environment with all required
tools using the `build-docker-dev` target.

```shell
make build-docker-dev
```

```shell
make build-docker-debug
```

```shell
docker run --rm -it --network host -u $(id -u) --env-file .env \
    -v .:/src go-ve/adan-bot:dev
```

Sharing Go build and modules cache with the container is easy, just mount some
extra volumes.

```shell
docker run --rm -it --network host -u $(id -u) --env-file .env \
    -v "$HOME/.cache:/.cache" -v "$HOME/go/pkg:/go/pkg" -v .:/src \
    go-ve/adan-bot:dev
```

This is equivalent to run the `dev-env` target.

```shell
make dev-env
```

**Variables:**

- `DOCKER_IMAGE`: Docker image name. (default: `go-ve/adan-bot`)
- `UID`: User ID that will run the binary. (default: current user ID)
- `ENV_FILE`: Populates the container environment with an env file. (default:
  `.env`)

### Hot reloading

#### air

If this is the first time you run this command in the project, we proceed to do the following:

```shell
make air-init
```

Then:

```shell
make air
```

#### watch

```shell
make watch
```

#### Variables

- `WATCH_TARGET`: Re-run given target. (default: `run`)

### Testing

```shell
make test
make test-race  # Enable the race condition detector during tests.
```

For generating coverage statistics you may use the `coverage` or `coverage-web`
targets, they will generate text and HTML outputs respectively.

```shell
make coverage
make coverage-web
```

Fuzz testing is also supported.

```shell
make fuzz
```

As well as benchmarks.

```shell
make benchmark
```

**Variables:**

- `GO`: Go toolchain to use. (default: `go`)
- `BENCHMARK_FILE`: Results file path. (default: `benchmarks-dev.txt`)
- `COVERAGE_FILE`: Coverage file path. (default: `coverage-dev.txt`)
- `CPUPROFILE`: CPU profile path. (default: `cpu.prof`)
- `MEMPROFILE`: Memory allocation profile path. (default: `mem.prof`)
- `TARGET_FUNC`: Run test that matches the given pattern. (default: `.`)
- `TARGET_PKG`: Run test on the given package. (default: `./...`)

### QA

```shell
make lint
```

It is possible to format Go files directly with the `format` target.

```shell
make format
```

Static code analysis is also provided, the are 2 variations.

```shell
make ca
make ca-fast  # Perform simple code analysis, reducing resources usage
```
