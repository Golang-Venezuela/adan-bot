ARG GO_VERSION=1.26.1

FROM golang:${GO_VERSION}-alpine3.23

RUN apk add --no-cache ca-certificates gcc git make musl-dev

WORKDIR /src

RUN go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install golang.org/x/perf/cmd/benchstat@latest && \
    go install golang.org/x/tools/cmd/godoc@latest && \
    go install github.com/air-verse/air@latest && \
    go clean -cache -modcache

VOLUME /.cache
VOLUME /go/pkg
VOLUME /src

CMD [ "air", "-c", ".air.toml" ]
