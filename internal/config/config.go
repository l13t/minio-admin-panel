package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	MinIOHost      string
	MinIOPort      int
	MinIOUseSSL    bool
	JWTSecret      string
	SessionTimeout int // in minutes
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("MINIO_PORT", "9000"))

	return &Config{
		MinIOHost:      getEnv("MINIO_HOST", "localhost"),
		MinIOPort:      port,
		MinIOUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		SessionTimeout: 60, // 1 hour
	}
}

// GetMinIOEndpoint returns the complete MinIO endpoint
func (c *Config) GetMinIOEndpoint() string {
	return fmt.Sprintf("%s:%d", c.MinIOHost, c.MinIOPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
