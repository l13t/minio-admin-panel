# GoReleaser compatible Dockerfile
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates=20240705-r0

WORKDIR /root/

# Copy the binary from GoReleaser
COPY minio-admin-panel .

# Copy web assets
COPY web ./web

# Copy environment example
COPY .env.example .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./minio-admin-panel"]
