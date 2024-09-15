# syntax=docker/dockerfile:1.9.0

FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG MODULE=github.com/betterde/cdns
ARG BUILD_VERSION=latest
ARG BINARY_NAME=cdns
ARG INSTALL_PATH=/usr/local/bin

RUN apk add --update gcc git make

ENV GOPATH=/tmp/buildcache
COPY . /go/src/cdns
WORKDIR /go/src/cdns
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X '${MODULE}/cmd.version=${BUILD_VERSION}' -X '${MODULE}/cmd.build=`date -u`' -X '${MODULE}/cmd.commit=`git rev-parse HEAD`'" -o bin/${BINARY_NAME} main.go

FROM --platform=$BUILDPLATFORM alpine:latest

ARG BUILD_VERSION=latest
ENV VERSION=${BUILD_VERSION}

LABEL org.opencontainers.image.url='https://github.com/betterde/cdns' \
      org.opencontainers.image.titile='CDNS' \
      org.opencontainers.image.vendor='Betterde Inc.' \
      org.opencontainers.image.source='https://github.com/betterde/cdns' \
      org.opencontainers.image.version=${VERSION} \
      org.opencontainers.image.authors='George <george@betterde.com>' \
      org.opencontainers.image.licenses='MIT' \
      org.opencontainers.image.description='An open-source lightweight DNS server that switches to ACME DNS challenge.' \
      org.opencontainers.image.documentation='https://github.com/betterde/cdns'

RUN apk --no-cache add curl ca-certificates \
    && update-ca-certificates

COPY --from=builder /go/src/cdns/bin/cdns /usr/local/bin/cdns
RUN mkdir -p /etc/cdns
WORKDIR /root/

HEALTHCHECK --interval=15s --timeout=3s \
  CMD curl -k -f https://127.0.0.1:443/health || exit 1

VOLUME ["/etc/cdns"]
ENTRYPOINT ["/usr/local/bin/cdns"]
EXPOSE 53 80 443
EXPOSE 53/udp
