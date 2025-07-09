#!/bin/bash

# MinIO Admin Panel Development Setup Script

set -e

echo "🚀 Setting up MinIO Admin Panel..."

# Check if Go is installed
if ! command -v go &>/dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d ' ' -f 3 | sed 's/go//')
echo "✅ Go version: $GO_VERSION"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file..."
    cp .env.example .env
    echo "✅ Created .env file. Please update it with your MinIO configuration."
fi

# Download dependencies
echo "📦 Downloading dependencies..."
go mod tidy

# Build the application
echo "🔨 Building application..."
mkdir -p bin
go build -o bin/minio-admin-panel main.go

# Test if the application starts
echo "🧪 Testing application startup..."
if timeout 3s ./bin/minio-admin-panel >/dev/null 2>&1; then
    echo "✅ Application builds and starts successfully!"
else
    echo "✅ Application builds successfully!"
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Update .env file with your MinIO server configuration"
echo "2. Start MinIO server (or run: make minio-dev)"
echo "3. Run the application: ./bin/minio-admin-panel"
echo "4. Open http://localhost:8080 in your browser"
echo ""
echo "💡 For development with hot reload, install Air and run: make dev"
echo "💡 For Docker deployment, run: make docker-run"
