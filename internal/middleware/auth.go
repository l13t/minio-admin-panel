package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username    string          `json:"username"`
	Password    string          `json:"password"` // Store encrypted password
	PolicyName  string          `json:"policy_name"`
	Permissions map[string]bool `json:"permissions"`
	jwt.RegisteredClaims
}

const JWTSecret = "your-secret-key" // Should come from config

// AuthRequired middleware checks for valid JWT token
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[DEBUG] Auth middleware checking token for %s %s", c.Request.Method, c.Request.URL.Path)

		// Check for token in cookie first
		tokenString, err := c.Cookie("token")
		if err != nil {
			log.Printf("[DEBUG] No token found in cookie, checking Authorization header")
			// Check Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				log.Printf("[DEBUG] No valid authorization found, redirecting to login")
				c.Redirect(http.StatusFound, "/")
				c.Abort()
				return
			}
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			log.Printf("[DEBUG] Found token in Authorization header")
		} else {
			log.Printf("[DEBUG] Found token in cookie")
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTSecret), nil
		})

		if err != nil || !token.Valid {
			log.Printf("[DEBUG] Token validation failed: %v", err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(*Claims); ok {
			log.Printf("[DEBUG] Token validated successfully for user '%s'", claims.Username)
			c.Set("username", claims.Username)
			c.Set("password", claims.Password)
			c.Set("policy_name", claims.PolicyName)
			c.Set("permissions", claims.Permissions)
			c.Set("user_claims", claims)
		} else {
			log.Printf("[DEBUG] Failed to extract claims from token")
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateJWT creates a JWT token for authenticated user
func GenerateJWT(username string) (string, error) {
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}

// GenerateJWTWithUserInfo creates a JWT token with user information and permissions
func GenerateJWTWithUserInfo(username, password, policyName string, permissions map[string]bool) (string, error) {
	claims := Claims{
		Username:    username,
		Password:    password, // In production, this should be encrypted
		PolicyName:  policyName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}

// GetUserPermissions extracts user permissions from context
func GetUserPermissions(c *gin.Context) map[string]bool {
	if permissions, exists := c.Get("permissions"); exists {
		if perms, ok := permissions.(map[string]bool); ok {
			return perms
		}
	}
	return map[string]bool{}
}

// CheckPermission checks if user has specific permission
func CheckPermission(c *gin.Context, permission string) bool {
	permissions := GetUserPermissions(c)
	return permissions[permission]
}

// RequirePermission middleware that requires specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !CheckPermission(c, permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
