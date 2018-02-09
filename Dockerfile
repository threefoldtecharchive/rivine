FROM golang:1.6.3
MAINTAINER rivine.io

ENV CGO_ENABLED 0
WORKDIR /go/src/github.com/rivine/rivine

RUN apt-get update && apt-get install -y zip

ENTRYPOINT ./release.sh
