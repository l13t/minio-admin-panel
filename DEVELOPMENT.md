# Development Guide

## Quick Start

1. **Run setup script**:

   ```bash
   ./setup.sh
   ```

2. **Start MinIO server** (if you3. **Testing with curl**:

   ```bash
   # Login with MinIO admin credentials
   curl -X POST http://localhost:8080/login \
     -d "username=minioadmin&password=minioadmin"
   
   # List buckets (with auth cookie)
   curl -X GET http://localhost:8080/buckets \
     -H "Accept: application/json" \
     --cookie-jar cookies.txt
   ```one running):

   ```bash
   make minio-dev
   ```

3. **Configure environment** (edit `.env`):

   ```bash
   MINIO_HOST=localhost
   MINIO_PORT=9000
   MINIO_USE_SSL=false
   JWT_SECRET=your-super-secret-jwt-key
   PORT=8080
   ```

4. **Run the application**:

   ```bash
   ./bin/minio-admin-panel
   ```

5. **Access the application**:
   - Open <http://localhost:8080>
   - Login with your MinIO credentials (default: minioadmin/minioadmin)

## Development Commands

```bash
# Build the application
make build

# Run the application
make run

# Run tests
make test

# Clean build artifacts
make clean

# Start MinIO for development
make minio-dev

# Stop MinIO development container
make minio-stop

# Run with Docker Compose
make docker-run

# Stop Docker Compose
make docker-stop
```

## Project Structure

```
internal/
├── config/      # Configuration management
├── handlers/    # HTTP request handlers
├── middleware/  # Authentication and other middleware
└── services/    # Business logic and MinIO integration

web/
├── templates/   # HTML templates
└── static/      # CSS, JavaScript, images
```

## API Endpoints

### Authentication

- `GET /` - Login page
- `POST /login` - Authenticate user
- `POST /logout` - Logout user

### Dashboard

- `GET /dashboard` - Main dashboard

### Buckets

- `GET /buckets` - List/manage buckets
- `POST /buckets` - Create bucket
- `DELETE /buckets/:name` - Delete bucket

### Users

- `GET /users` - List/manage users
- `POST /users` - Create user
- `DELETE /users/:name` - Delete user

## Troubleshooting

### Common Issues

1. **"Connection refused" error**:
   - Make sure MinIO server is running
   - Check `MINIO_HOST` and `MINIO_PORT` in `.env`

2. **"Authentication failed"**:
   - Verify MinIO credentials in `.env`
   - Ensure the user has admin privileges

3. **"Template not found"**:
   - Ensure you're running from the project root
   - Check that `web/templates/` directory exists

### Development Tips

1. **Hot reload**: Install `air` for automatic rebuilds:

   ```bash
   go install github.com/cosmtrek/air@latest
   make dev
   ```

2. **Debug logging**: Set `GIN_MODE=debug` for verbose logging

3. **Testing with curl**:

   ```bash
   # Login
   curl -X POST http://localhost:8080/login \
     -d "username=admin&password=admin123"
   
   # List buckets (with auth cookie)
   curl -X GET http://localhost:8080/buckets \
     -H "Accept: application/json" \
     --cookie-jar cookies.txt
   ```

## Building for Production

```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o minio-admin-panel main.go

# Or use Docker
docker build -t minio-admin-panel .
```
