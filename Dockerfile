FROM golang:1.22-alpine AS builder
MAINTAINER George "george@betterde.com"

ARG MODULE=github.com/betterde/cdns
ARG VERSION=latest
ARG BINARY_NAME=cdns
ARG INSTALL_PATH=/usr/local/bin

RUN apk add --update gcc git make

ENV GOPATH /tmp/buildcache
COPY . /go/src/cdns
WORKDIR /go/src/cdns
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X '${MODULE}/cmd.version=${VERSION}' -X '${MODULE}/cmd.build=`date -u`' -X '${MODULE}/cmd.commit=`git rev-parse HEAD`'" -o bin/${BINARY_NAME} main.go

FROM alpine:latest
MAINTAINER George "george@betterde.com"
WORKDIR /root/
COPY --from=builder /go/src/cdns/bin/cdns /usr/local/bin/cdns
RUN mkdir -p /etc/cdns
COPY root_ca.crt /usr/local/share/ca-certificates/ca.crt
RUN apk --no-cache add ca-certificates && update-ca-certificates

VOLUME ["/etc/cdns"]
ENTRYPOINT ["/usr/local/bin/cdns"]
EXPOSE 53 80 443
EXPOSE 53/udp
