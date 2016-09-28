FROM golang:1.6.3
MAINTAINER rivine.io

ENV CGO_ENABLED 0
COPY . /go/src/github.com/rivine/rivine
WORKDIR /go/src/github.com/rivine/rivine

EXPOSE 23112

ENTRYPOINT go get -d -v ./... && go install ./... && rivined

