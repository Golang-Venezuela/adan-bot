ARG GO_VERSION=1.26.1

FROM golang:${GO_VERSION}-alpine3.23

RUN apk add --no-cache ca-certificates build-base binutils-gold git make

WORKDIR /src

RUN CGO_ENABLED=0 go install github.com/go-delve/delve/cmd/dlv@latest && \
    CGO_ENABLED=0 go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    CGO_ENABLED=0 go install golang.org/x/perf/cmd/benchstat@latest && \
    CGO_ENABLED=0 go install golang.org/x/tools/cmd/godoc@latest && \
    CGO_ENABLED=0 go install github.com/air-verse/air@latest && \
    go clean -cache -modcache

VOLUME /.cache
VOLUME /go/pkg
VOLUME /src

CMD [ "air", "-c", ".air.toml" ]
