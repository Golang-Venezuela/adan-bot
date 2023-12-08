ARG GO_VERSION=1.21

FROM golang:${GO_VERSION}-alpine3.18 AS builder

LABEL maintainer="Eduardo Bravo <eduardojosebb.matescience@gmail.com>"

RUN apk --no-cache add ca-certificates && update-ca-certificates
WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 GOFLAGS="-tags=timetzdata" \
  go build -ldflags="-s -w" -trimpath -o ./dist/ ./...

FROM alpine:3.18 as debug
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/dist/adan-bot /bin/adan-bot
USER 1000
CMD ["/bin/adan-bot"]

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/dist/adan-bot /bin/adan-bot
USER 1000
ENTRYPOINT ["/bin/adan-bot"]
