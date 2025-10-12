FROM golang:1.21-alpine AS builder

# Install step CLI
RUN apk add --no-cache curl
RUN curl -L https://github.com/smallstep/cli/releases/latest/download/step_linux_amd64.tar.gz | tar xz -C /usr/local/bin

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Final stage
FROM alpine:latest

# Install step CLI and openssl
RUN apk add --no-cache curl openssl
RUN curl -L https://github.com/smallstep/cli/releases/latest/download/step_linux_amd64.tar.gz | tar xz -C /usr/local/bin

# Create app user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Create data directory
RUN mkdir -p /app/data && chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
