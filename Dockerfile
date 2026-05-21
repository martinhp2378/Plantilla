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

# Run stage
FROM alpine:latest

# Install CA certificates for HTTPS requests (useful for private repos/APIs)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

# Copy necessary runtime assets
COPY --from=builder /app/sql ./sql
COPY --from=builder /app/static ./static

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]
