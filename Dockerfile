# Build stage
FROM golang:1.24.1-alpine AS builder

# Install build dependencies for CGO (required for SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" -o timesheet ./cmd/timesheet

# Runtime stage
FROM alpine:latest

# Install SQLite and CA certificates
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/timesheet .

# Create directory for database
RUN mkdir -p /app/data && chmod 755 /app/data

# Expose port
EXPOSE 8080

# Run the API server only (no TUI)
CMD ["./timesheet", "--no-tui", "--port", "8080"]

