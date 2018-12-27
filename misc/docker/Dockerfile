FROM golang:1.11-alpine AS base

RUN apk add --no-cache git bash ca-certificates build-base
RUN go get -u github.com/derekparker/delve/cmd/dlv


FROM alpine:3.8

COPY ./unix-logger /usr/local/bin/unix-logger
ENTRYPOINT /usr/local/bin/unix-logger