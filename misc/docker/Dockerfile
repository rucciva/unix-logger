FROM golang:1.11-alpine AS base

RUN apk add --no-cache git bash ca-certificates build-base
RUN go get -u github.com/derekparker/delve/cmd/dlv


FROM alpine:3.8

COPY ./unix-logger /usr/local/bin/unix-logger

ENV UNIX_LOGGER_PATH ""
ENV UNIX_LOGGER_FILE_OWNER ""
ENV UNIX_LOGGER_FILE_MODE ""
ENV UNIX_LOGGER_MAX_CONNECTION ""

ENTRYPOINT /usr/local/bin/unix-logger