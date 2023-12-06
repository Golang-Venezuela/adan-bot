FROM golang:1.21-alpine3.18
RUN apk add --no-cache ca-certificates gcc make musl-dev
RUN \
  go install github.com/cespare/reflex@v0.3.1 && \
  go install github.com/go-delve/delve/cmd/dlv@v1.21.1 && \
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2 && \
  go install golang.org/x/perf/cmd/benchstat@master && \
  go install golang.org/x/tools/cmd/godoc@master && \
  go clean -cache -modcache
WORKDIR /src
VOLUME /.cache
VOLUME /go/pkg
VOLUME /src
