package handlers

import (
	"log"
	"net/http"

	"minio-admin-panel/internal/middleware"
	"minio-admin-panel/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	minioService *services.MinIOService
}

func NewAuthHandler(minioService *services.MinIOService) *AuthHandler {
	return &AuthHandler{
		minioService: minioService,
	}
}

// LoginPage renders the login form
func (h *AuthHandler) LoginPage(c *gin.Context) {
	// Check if user is already authenticated
	if token, err := c.Cookie("token"); err == nil && token != "" {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	RenderWithTranslations(c, "login.html", gin.H{
		"title": "login.title",
	})
}

// Login handles authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var loginData struct {
		Username string `form:"username" json:"username" binding:"required"`
		Password string `form:"password" json:"password" binding:"required"`
	}

	if err := c.ShouldBind(&loginData); err != nil {
		log.Printf("[DEBUG] Login form validation failed: %v", err)
		RenderWithTranslations(c, "login.html", gin.H{
			"title": "login.title",
			"error": "login.error.missing_credentials",
		})
		return
	}

	log.Printf("[DEBUG] Login attempt for user '%s' from IP %s", loginData.Username, c.ClientIP())

	// Validate credentials with admin panel credentials
	userInfo, err := h.minioService.ValidateCredentials(loginData.Username, loginData.Password)
	if err != nil {
		log.Printf("[DEBUG] Login failed for user '%s': %v", loginData.Username, err)
		RenderWithTranslations(c, "login.html", gin.H{
			"title": "login.title",
			"error": "login.error.invalid_credentials",
		})
		return
	}

	log.Printf("[DEBUG] Credential validation successful for user '%s'", loginData.Username)

	// Get user permissions
	permissions := h.minioService.GetUserPermissions(loginData.Username, loginData.Password)
	log.Printf("[DEBUG] Retrieved permissions for user '%s': %+v", loginData.Username, permissions)

	// Generate JWT token with user info
	token, err := middleware.GenerateJWTWithUserInfo(loginData.Username, loginData.Password, userInfo.PolicyName, permissions)
	if err != nil {
		log.Printf("[DEBUG] JWT token generation failed for user '%s': %v", loginData.Username, err)
		RenderWithTranslations(c, "login.html", gin.H{
			"title": "login.title",
			"error": "login.error.token_generation",
		})
		return
	}

	log.Printf("[DEBUG] JWT token generated successfully for user '%s'", loginData.Username)

	// Set cookie
	c.SetCookie("token", token, 3600, "/", "", false, true)

	log.Printf("[DEBUG] Login successful for user '%s', redirecting to dashboard", loginData.Username)
	c.Redirect(http.StatusFound, "/dashboard")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}
