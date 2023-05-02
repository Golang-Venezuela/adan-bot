ARG GO_VERSION=1.19

FROM golang:${GO_VERSION}-alpine AS builder

LABEL maintainer="Eduardo Bravo <eduardojosebb.matescience@gmail.com>"

RUN go env -w GOPROXY=direct
RUN apk add --no-cache git
RUN apk --no-cache add ca-certificates && update-ca-certificates

WORKDIR /adan-bot/

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
     -a -installsuffix cgo -o main .

FROM alpine AS bot-runner

WORKDIR /adan-bot/
COPY --from=builder /adan-bot/main  .
COPY --from=builder /adan-bot/.env  .

CMD ["./main"]


