# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy the entire source code (including vendor directory)
COPY . .

# Install templ for generation
RUN go install github.com/a-h/templ/cmd/templ@latest

# Generate Templ files
RUN templ generate

# Build the application using the vendor directory
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o main ./cmd/main.go

# Development stage with live reload (Air + Templ)
FROM builder AS dev
RUN go install github.com/air-verse/air@latest
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# Run stage
FROM alpine:latest AS runtime

# Install CA certificates for HTTPS requests (useful for private repos/APIs)
RUN apk add --no-cache ca-certificates

# Create system group and user with UID 10001
RUN addgroup -g 10001 -S appuser && adduser -u 10001 -S -G appuser appuser

WORKDIR /app

# Copy the binary from builder with correct ownership
COPY --chown=appuser:appuser --from=builder /app/main .

# Copy necessary runtime assets with correct ownership
COPY --chown=appuser:appuser --from=builder /app/sql ./sql
COPY --chown=appuser:appuser --from=builder /app/static ./static

# Switch to non-root user
USER 10001

# Expose port
EXPOSE 8080

# Health check against lightweight liveness endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

# Command to run
CMD ["./main"]
