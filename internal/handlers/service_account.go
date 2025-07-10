package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"minio-admin-panel/internal/services"

	"github.com/gin-gonic/gin"
)

type ServiceAccountHandler struct {
	minioService *services.MinIOService
}

func NewServiceAccountHandler(minioService *services.MinIOService) *ServiceAccountHandler {
	return &ServiceAccountHandler{
		minioService: minioService,
	}
}

// Helper function to extract credentials from context
func (h *ServiceAccountHandler) getCredentials(c *gin.Context) (string, string, error) {
	username, usernameExists := c.Get("username")
	password, passwordExists := c.Get("password")

	if !usernameExists || !passwordExists || username == nil || password == nil {
		return "", "", fmt.Errorf("missing credentials")
	}

	return username.(string), password.(string), nil
}

// ListServiceAccounts handles GET /api/service-accounts?user=<username>
func (h *ServiceAccountHandler) ListServiceAccounts(c *gin.Context) {
	log.Printf("[DEBUG] ListServiceAccounts request from %s %s", c.Request.Method, c.Request.URL.Path)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in ListServiceAccounts: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	targetUser := c.Query("user")
	if targetUser == "" {
		log.Printf("[DEBUG] Missing target user parameter in ListServiceAccounts")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user parameter"})
		return
	}

	log.Printf("[DEBUG] ListServiceAccounts for target user '%s' by admin '%s'", targetUser, username)
	serviceAccounts, err := h.minioService.ListServiceAccounts(context.Background(), targetUser, username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to list service accounts for user '%s': %v", targetUser, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list service accounts: %v", err)})
		return
	}

	log.Printf("[DEBUG] ListServiceAccounts successful for user '%s', returning %d service accounts", targetUser, len(serviceAccounts))
	c.JSON(http.StatusOK, gin.H{"service_accounts": serviceAccounts})
}

// CreateServiceAccount handles POST /api/service-accounts
func (h *ServiceAccountHandler) CreateServiceAccount(c *gin.Context) {
	log.Printf("[DEBUG] CreateServiceAccount request from %s %s", c.Request.Method, c.Request.URL.Path)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in CreateServiceAccount: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	var req struct {
		TargetUser  string `json:"target_user" binding:"required"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[DEBUG] Failed to bind JSON in CreateServiceAccount: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body: %v", err)})
		return
	}

	log.Printf("[DEBUG] CreateServiceAccount for target user '%s' with name '%s' by admin '%s'", req.TargetUser, req.Name, username)
	serviceAccount, err := h.minioService.CreateServiceAccount(context.Background(), req.TargetUser, req.Name, req.Description, username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create service account for user '%s': %v", req.TargetUser, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create service account: %v", err)})
		return
	}

	log.Printf("[DEBUG] CreateServiceAccount successful for user '%s', access key: '%s'", req.TargetUser, serviceAccount["access_key"])
	c.JSON(http.StatusOK, gin.H{"service_account": serviceAccount})
}

// DeleteServiceAccount handles DELETE /api/service-accounts/:accessKey
func (h *ServiceAccountHandler) DeleteServiceAccount(c *gin.Context) {
	log.Printf("[DEBUG] DeleteServiceAccount request from %s %s", c.Request.Method, c.Request.URL.Path)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in DeleteServiceAccount: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	accessKey := c.Param("accessKey")
	if accessKey == "" {
		log.Printf("[DEBUG] Missing access key parameter in DeleteServiceAccount")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing access key parameter"})
		return
	}

	log.Printf("[DEBUG] DeleteServiceAccount for access key '%s' by admin '%s'", accessKey, username)
	err = h.minioService.DeleteServiceAccount(context.Background(), accessKey, username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to delete service account '%s': %v", accessKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete service account: %v", err)})
		return
	}

	log.Printf("[DEBUG] DeleteServiceAccount successful for access key '%s'", accessKey)
	c.JSON(http.StatusOK, gin.H{"message": "Service account deleted successfully"})
}

// GetServiceAccountInfo handles GET /api/service-accounts/:accessKey
func (h *ServiceAccountHandler) GetServiceAccountInfo(c *gin.Context) {
	log.Printf("[DEBUG] GetServiceAccountInfo request from %s %s", c.Request.Method, c.Request.URL.Path)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in GetServiceAccountInfo: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	accessKey := c.Param("accessKey")
	if accessKey == "" {
		log.Printf("[DEBUG] Missing access key parameter in GetServiceAccountInfo")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing access key parameter"})
		return
	}

	log.Printf("[DEBUG] GetServiceAccountInfo for access key '%s' by admin '%s'", accessKey, username)
	serviceAccount, err := h.minioService.GetServiceAccountInfo(context.Background(), accessKey, username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to get service account info '%s': %v", accessKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get service account info: %v", err)})
		return
	}

	log.Printf("[DEBUG] GetServiceAccountInfo successful for access key '%s'", accessKey)
	c.JSON(http.StatusOK, gin.H{"service_account": serviceAccount})
}
