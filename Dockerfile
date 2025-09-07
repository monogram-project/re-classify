# Build stage
FROM docker.io/golang:1.25-alpine AS builder

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
FROM docker.io/alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/re-classify .

# Create a non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser

ENTRYPOINT ["./re-classify"]

