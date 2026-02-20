# Build from repo root: docker build -t lazeez-core .
# Run in cloud: set env (PORT, DATABASE_URL or DB_*, JWT_SECRET_KEY, CORS_ALLOWED_ORIGINS, etc.)

# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/main ./cmd

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates wget

# Non-root user for security in cloud
RUN adduser -D -g "" appuser
WORKDIR /app

COPY --from=builder /app/main .
RUN chown appuser:appuser /app/main

USER appuser

EXPOSE 8080

# Health check for orchestrators (K8s, ECS, etc.). Uses PORT env when set.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD sh -c 'wget -q -O- "http://127.0.0.1:${PORT:-8080}/health" || exit 1'

CMD ["./main"]
