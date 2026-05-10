# syntax=docker/dockerfile:1

# Pin Go minor version; bump if go.mod requires a newer toolchain.
ARG GO_VERSION=1.25

# -----------------------------------------------------------------------------
# Stage 1: download modules (cached unless go.mod / go.sum change)
# -----------------------------------------------------------------------------
FROM golang:${GO_VERSION}-alpine AS deps
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download

# -----------------------------------------------------------------------------
# Stage 2: build static binary
# -----------------------------------------------------------------------------
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates git

ENV GOTOOLCHAIN=auto
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY go.mod go.sum ./
COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath \
    -ldflags="-s -w" \
    -o /out/collector \
    ./cmd/collector

# -----------------------------------------------------------------------------
# Stage 3: minimal runtime
# -----------------------------------------------------------------------------
FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 65532 -S app \
    && adduser -u 65532 -S -G app -h /nonexistent -s /sbin/nologin app

WORKDIR /app

COPY --from=builder /out/collector /app/collector

USER app:app
ENTRYPOINT ["/app/collector"]
