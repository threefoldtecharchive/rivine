FROM golang:1.9
MAINTAINER  threefold.tech

ENV CGO_ENABLED 0
COPY . /go/src/github.com/threefoldtech/rivine
WORKDIR /go/src/github.com/threefoldtech/rivine

EXPOSE 23110 23112

RUN go install -v -tags 'debug dev profile' ./... 
ENTRYPOINT ["rivined"]
CMD ["--no-bootstrap"]