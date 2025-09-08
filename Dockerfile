# GoReleaser-compatible Dockerfile
# This expects a pre-built binary to be copied in
FROM docker.io/alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy the pre-built binary (provided by GoReleaser)
COPY re-classify .

# Set execute permissions on the binary
RUN chmod +x re-classify

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

# Change ownership of the app directory
RUN chown -R appuser:appuser /app

USER appuser

ENTRYPOINT ["/app/re-classify"]

