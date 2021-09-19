#syntax=docker/dockerfile:1.2
FROM golang:alpine AS build

ARG GO_OS=linux
ARG GO_ARCH=amd64
ARG CGO_ENABLED=0

RUN apk --update add ca-certificates git

WORKDIR /build

COPY . /build/

RUN go mod tidy
RUN GOOS=$GO_OS CGO_ENABLED=$CGO_ENABLED go build -o rbn cmd/cli/main.go

FROM alpine AS runtime

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /app

COPY --from=build /build/rbn /app

ENTRYPOINT ["/app/rbn"]
