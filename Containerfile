FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o complypack ./cmd/complypack

FROM registry.access.redhat.com/ubi9-micro:latest

COPY --from=builder /build/complypack /usr/local/bin/complypack

ENTRYPOINT ["complypack"]
