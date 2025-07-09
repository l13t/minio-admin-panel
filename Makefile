.PHONY: build run clean test docker-build docker-run dev

# Build the application
build:
	go build -o bin/minio-admin-panel main.go

# Run the application
run:
	go run main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test -v ./...

# Test authentication (requires running server)
test-auth:
	./test_auth.sh

# Download dependencies
deps:
	go mod download
	go mod tidy

# Development mode with air for hot reload
dev:
	air

# Docker build
docker-build:
	docker build -t minio-admin-panel .

# Docker run
docker-run:
	docker-compose up -d

# Docker stop
docker-stop:
	docker-compose down

# Start MinIO only for development
minio-dev:
	docker run -d \
		--name minio-dev \
		-p 9000:9000 \
		-p 9001:9001 \
		-e "MINIO_ROOT_USER=minioadmin" \
		-e "MINIO_ROOT_PASSWORD=minioadmin" \
		minio/minio server /data --console-address ":9001"

# Stop MinIO dev container
minio-stop:
	docker stop minio-dev || true
	docker rm minio-dev || true

# Setup development environment
setup:
	cp .env.example .env
	go mod tidy

# Validate configuration
validate:
	go run cmd/validate/main.go
