FROM golang:1.19 as golang

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /build
ADD . /build

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -o linguaDB

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache bash
COPY --from=golang /build/linguaDB /app
COPY --from=golang /build/config.yaml /app
EXPOSE 8000
ENTRYPOINT ["/app/linguaDB"]