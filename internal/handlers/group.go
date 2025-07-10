package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"minio-admin-panel/internal/services"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	minioService *services.MinIOService
}

func NewGroupHandler(minioService *services.MinIOService) *GroupHandler {
	return &GroupHandler{
		minioService: minioService,
	}
}

// Helper function to extract credentials from context
func (h *GroupHandler) getCredentials(c *gin.Context) (string, string, error) {
	username, usernameExists := c.Get("username")
	password, passwordExists := c.Get("password")

	if !usernameExists || !passwordExists || username == nil || password == nil {
		return "", "", fmt.Errorf("missing credentials")
	}

	return username.(string), password.(string), nil
}

// ListGroups handles GET /groups
func (h *GroupHandler) ListGroups(c *gin.Context) {
	log.Printf("[DEBUG] ListGroups request from %s %s", c.Request.Method, c.Request.URL.Path)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in ListGroups: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] ListGroups for admin '%s'", username)
	groups, err := h.minioService.ListGroups(context.Background(), username, password)
	if err != nil {
		log.Printf("[DEBUG] ListGroups failed for admin '%s': %v", username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] ListGroups successful for admin '%s', found %d groups", username, len(groups))

	// Check if this is an API request
	if c.GetHeader("Accept") == "application/json" {
		log.Printf("[DEBUG] Returning JSON response with %d groups", len(groups))
		c.JSON(http.StatusOK, gin.H{"groups": groups})
		return
	}

	// Render HTML page
	log.Printf("[DEBUG] Rendering HTML page with %d groups", len(groups))
	permissions, _ := c.Get("permissions")
	policyName, _ := c.Get("policy_name")

	RenderWithTranslations(c, "groups.html", gin.H{
		"title":       "groups.title",
		"groups":      groups,
		"permissions": permissions,
		"username":    username,
		"policy_name": policyName,
	})
}

// CreateGroup handles POST /groups
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	username, password, err := h.getCredentials(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	var req struct {
		Name string `form:"name" json:"name" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.minioService.CreateGroup(context.Background(), req.Name, username, password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Group created successfully"})
}

// DeleteGroup handles DELETE /groups/:name
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	username, password, err := h.getCredentials(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	groupName := c.Param("name")

	if err := h.minioService.DeleteGroup(context.Background(), groupName, username, password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

// GetGroupInfo handles GET /groups/:name
func (h *GroupHandler) GetGroupInfo(c *gin.Context) {
	username, password, err := h.getCredentials(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	groupName := c.Param("name")

	groupInfo, err := h.minioService.GetGroupInfo(context.Background(), groupName, username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groupInfo)
}

// SetGroupPolicy handles PUT /groups/:name/policy
func (h *GroupHandler) SetGroupPolicy(c *gin.Context) {
	username, password, err := h.getCredentials(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	groupName := c.Param("name")

	var req struct {
		PolicyName string `form:"policy_name" json:"policy_name" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.minioService.SetGroupPolicy(context.Background(), groupName, req.PolicyName, username, password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group policy updated successfully"})
}

// UpdateGroupMembers handles PUT /groups/:name/members
func (h *GroupHandler) UpdateGroupMembers(c *gin.Context) {
	username, password, err := h.getCredentials(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	groupName := c.Param("name")

	var req struct {
		AddUsers    []string `form:"add_users" json:"add_users"`
		RemoveUsers []string `form:"remove_users" json:"remove_users"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Add users to group
	if len(req.AddUsers) > 0 {
		if err := h.minioService.AddUsersToGroup(context.Background(), groupName, req.AddUsers, username, password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add users: %v", err)})
			return
		}
	}

	// Remove users from group
	if len(req.RemoveUsers) > 0 {
		if err := h.minioService.RemoveUsersFromGroup(context.Background(), groupName, req.RemoveUsers, username, password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to remove users: %v", err)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group members updated successfully"})
}

// SetUserGroups handles PUT /users/:name/groups - sets all groups for a user
func (h *GroupHandler) SetUserGroups(c *gin.Context) {
	username, password, err := h.getCredentials(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	userName := c.Param("name")

	var req struct {
		Groups []string `form:"groups" json:"groups"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get current user info to see existing groups
	userInfo, err := h.minioService.GetUser(context.Background(), userName, username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get user info: %v", err)})
		return
	}

	// Remove user from all current groups
	if len(userInfo.MemberOf) > 0 {
		for _, groupName := range userInfo.MemberOf {
			if err := h.minioService.RemoveUsersFromGroup(context.Background(), groupName, []string{userName}, username, password); err != nil {
				log.Printf("[DEBUG] Failed to remove user '%s' from group '%s': %v", userName, groupName, err)
				// Continue to try removing from other groups
			}
		}
	}

	// Add user to new groups
	for _, groupName := range req.Groups {
		if strings.TrimSpace(groupName) != "" {
			if err := h.minioService.AddUsersToGroup(context.Background(), groupName, []string{userName}, username, password); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add user to group '%s': %v", groupName, err)})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "User groups updated successfully"})
}
