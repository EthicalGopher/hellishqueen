# Debug Dockerfile - let's see what's happening
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Debug: Show what's in the build context
RUN echo "=== Initial directory ===" && ls -la

# Copy go mod files
COPY go.mod go.sum ./

# Debug: Show go mod files
RUN echo "=== After copying go.mod ===" && ls -la && cat go.mod

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Debug: Show all files after copying source
RUN echo "=== After copying source ===" && ls -la

# Debug: Try to find main.go
RUN find . -name "*.go" -type f

# Debug: Check Go version and environment
RUN go version && go env

# Try building with verbose output
RUN echo "=== Building ===" && go build -v -o app .

# Debug: Check if binary was created
RUN echo "=== After build ===" && ls -la && file app || echo "app not found"

# ---

# Simple runtime stage for testing
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
RUN ls -la && file app 2>/dev/null || echo "Binary not found in runtime"
CMD ["./app"]