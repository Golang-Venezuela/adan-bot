ARG GO_VERSION=1.24

FROM golang:${GO_VERSION}-alpine3.22
RUN apk add --no-cache ca-certificates gcc git make musl-dev
RUN \
  go install github.com/go-delve/delve/cmd/dlv@v1.21.1 && \
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
RUN \
  go install golang.org/x/perf/cmd/benchstat@master && \
  go install golang.org/x/tools/cmd/godoc@master
RUN \
  go install github.com/air-verse/air@latest
RUN go clean -cache -modcache

WORKDIR /src
VOLUME /.cache
VOLUME /go/pkg
VOLUME /src

CMD [ "air", "-c", ".air.toml" ]
