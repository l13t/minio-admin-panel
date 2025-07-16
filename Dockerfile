FROM golang:1.24 AS builder

# Set the working directory
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./
# Copy the source code
COPY . .
# Download depended

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -o minio-admin-panel .

# Use a minimal base image for the final binary
FROM scratch

COPY --from=builder /app/minio-admin-panel /minio-admin-panel
COPY --from=builder /app/translations /translations
COPY --from=builder /app/web /web

# Set the entrypoint for the container
ENTRYPOINT ["/minio-admin-panel"]
# Expose the port the app runs on
EXPOSE 8080
