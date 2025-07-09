# MinIO Admin Panel

A modern web interface for managing MinIO object storage instances. This Go-based web application provides a user-friendly dashboard for managing buckets, users, policies, and monitoring your MinIO deployment.

## Features

- 🔐 **Secure Authentication** - Login using MinIO admin credentials directly
- 🛡️ **Admin Access Control** - Validates admin privileges against MinIO server
- 📊 **Dashboard** - Overview of your MinIO instance with key metrics
- 🪣 **Bucket Management** - Create, delete, and manage bucket policies
- 👥 **User Management** - Create, delete users and manage their policies
- 🎨 **Modern UI** - Clean, responsive interface built with Bootstrap
- 🚀 **Fast & Lightweight** - Built with Go and Gin framework
- 🐳 **Docker Ready** - Easy deployment with Docker and Docker Compose

## Screenshots

### Login Page

Clean authentication interface that validates admin credentials against MinIO server.

### Dashboard

Overview of your MinIO instance with key metrics and quick actions.

### Bucket Management

Manage buckets with easy-to-use interface for creation, deletion, and policy management.

### User Management

Create and manage MinIO users with policy assignments.

## Quick Start

### Prerequisites

- Go 1.21 or later
- MinIO server running (for development, see [MinIO Setup](#minio-setup))

### Installation

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd minio-admin-panel
   ```

2. **Setup environment**

   ```bash
   make setup
   ```

3. **Configure environment variables**
   Edit `.env` file:

   ```bash
   MINIO_HOST=localhost
   MINIO_PORT=9000
   MINIO_USE_SSL=false
   JWT_SECRET=your-super-secret-jwt-key
   PORT=8080
   ```

4. **Run the application**

   ```bash
   make run
   ```

5. **Access the application**
   Open <http://localhost:8080> in your browser

## Development

### Using Make Commands

```bash
# Install dependencies
make deps

# Run in development mode
make dev

# Build the application
make build

# Run tests
make test

# Clean build artifacts
make clean
```

### MinIO Setup

For development, you can quickly start a MinIO instance:

```bash
# Start MinIO for development
make minio-dev

# Stop MinIO
make minio-stop
```

MinIO Console will be available at <http://localhost:9001>

### Docker Development

Use Docker Compose for a complete development environment:

```bash
# Start both MinIO and Admin Panel
make docker-run

# Stop services
make docker-stop
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MINIO_HOST` | MinIO server hostname or IP | `localhost` |
| `MINIO_PORT` | MinIO server port | `9000` |
| `MINIO_USE_SSL` | Use SSL for MinIO connection | `false` |
| `JWT_SECRET` | JWT secret for session management | `your-secret-key` |
| `PORT` | Server port | `8080` |

## API Endpoints

### Authentication

- `GET /` - Login page
- `POST /login` - Authenticate user
- `POST /logout` - Logout user

### Dashboard

- `GET /dashboard` - Dashboard page

### Buckets

- `GET /buckets` - List buckets
- `POST /buckets` - Create bucket
- `DELETE /buckets/:name` - Delete bucket
- `GET /buckets/:name/policy` - Get bucket policy
- `PUT /buckets/:name/policy` - Set bucket policy

### Users

- `GET /users` - List users
- `POST /users` - Create user
- `DELETE /users/:name` - Delete user
- `PUT /users/:name/policy` - Set user policy

### API Routes

- `GET /api/server-info` - Get server information
- `GET /api/metrics` - Get server metrics

## Project Structure

```
minio-admin-panel/
├── main.go                 # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # Middleware (auth, etc.)
│   └── services/          # Business logic
├── web/
│   ├── static/           # Static assets (CSS, JS, images)
│   └── templates/        # HTML templates
├── docker-compose.yml    # Docker Compose configuration
├── Dockerfile           # Docker image definition
├── Makefile            # Build and development commands
└── README.md          # This file
```

## Security

- **JWT Authentication**: Secure session management with JWT tokens
- **MinIO Validation**: All credentials are validated against MinIO server
- **HTTPS Support**: Configurable SSL/TLS support
- **Session Timeout**: Configurable session timeout
- **Input Validation**: Comprehensive input validation and sanitization

## User Authentication & Authorization

The MinIO Admin Panel provides secure access control by validating admin credentials directly against the MinIO server:

### Authentication Flow

1. User enters MinIO admin username and password in the login form
2. System validates credentials against MinIO server using admin API
3. System verifies that the user has admin privileges (can perform admin operations)
4. JWT token is generated for session management with user credentials
5. All subsequent MinIO operations use the authenticated user's credentials

### Admin Privileges Required

To access the admin panel, users must have MinIO admin privileges, which means they can:

- List and manage MinIO users
- Perform administrative operations
- Access all buckets and data
- Manage policies and configurations

### Security Features

- **Direct MinIO Authentication**: No separate admin panel credentials - uses actual MinIO admin credentials
- **Admin Privilege Validation**: Ensures only users with admin privileges can access the panel
- **JWT Session Management**: Secure session tokens with configurable expiration
- **Credential Validation**: Real-time validation against MinIO server
- **HTTPS Support**: Full SSL/TLS support for encrypted communications

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Production Deployment

### Docker

1. **Build and run with Docker Compose**

   ```bash
   docker-compose up -d
   ```

2. **Or build manually**

   ```bash
   docker build -t minio-admin-panel .
   docker run -d \
     -p 8080:8080 \
     -e MINIO_HOST=your-minio-server \
     -e MINIO_PORT=9000 \
     -e MINIO_USE_SSL=false \
     -e JWT_SECRET=your-production-jwt-secret \
     minio-admin-panel
   ```

### Binary

1. **Build for production**

   ```bash
   CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o minio-admin-panel main.go
   ```

2. **Run with environment variables**

   ```bash
   MINIO_HOST=your-minio-server \
   MINIO_PORT=9000 \
   MINIO_USE_SSL=false \
   JWT_SECRET=your-production-jwt-secret \
   ./minio-admin-panel
   ```

## Troubleshooting

### Common Issues

1. **Connection refused to MinIO**
   - Ensure MinIO server is running
   - Check the `MINIO_HOST` and `MINIO_PORT` configuration
   - Verify network connectivity

2. **Authentication failures**
   - Verify your MinIO admin username and password
   - Ensure the user has admin privileges in MinIO

3. **Template not found errors**
   - Ensure `web/templates/` directory exists and contains HTML files
   - Check file permissions

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [MinIO](https://min.io/) - High Performance Object Storage
- [Gin](https://gin-gonic.com/) - Go Web Framework
- [Bootstrap](https://getbootstrap.com/) - CSS Framework
