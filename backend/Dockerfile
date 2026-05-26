# =============================================================================
# Multi-stage build for all Go services
# =============================================================================

# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy source code
COPY . .

# Force cache bust on rebuild
ARG CACHEBUST=1

# Download dependencies and generate go.sum
RUN go mod tidy && go mod download

# Build API
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/api ./cmd/api

# Build DNS
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/dns ./cmd/dns

# Build SMTP
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/smtp ./cmd/smtp

# Build Worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/worker ./cmd/worker

# =============================================================================
# API Runtime
# =============================================================================
FROM alpine:3.19 AS api

RUN apk add --no-cache ca-certificates tzdata wget
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /bin/api /usr/local/bin/api

USER app
ENTRYPOINT ["/usr/local/bin/api"]

# =============================================================================
# DNS Runtime
# =============================================================================
FROM alpine:3.19 AS dns

RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /bin/dns /usr/local/bin/dns

USER app
ENTRYPOINT ["/usr/local/bin/dns"]

# =============================================================================
# SMTP Runtime
# =============================================================================
FROM alpine:3.19 AS smtp

RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /bin/smtp /usr/local/bin/smtp

USER app
ENTRYPOINT ["/usr/local/bin/smtp"]

# =============================================================================
# Worker Runtime
# =============================================================================
FROM alpine:3.19 AS worker

RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /bin/worker /usr/local/bin/worker

USER app
ENTRYPOINT ["/usr/local/bin/worker"]
