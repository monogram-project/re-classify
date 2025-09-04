# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o re-classify ./cmd/re-classify

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/re-classify .

# Copy test configs for examples
COPY --from=builder /app/test-configs ./test-configs

# Create a non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser

ENTRYPOINT ["./re-classify"]
