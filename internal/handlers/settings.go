package handlers

import (
	"minio-admin-panel/internal/services"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	minioService *services.MinIOService
	version      string
	commit       string
	date         string
	builtBy      string
}

func NewSettingsHandler(minioService *services.MinIOService, version, commit, date, builtBy string) *SettingsHandler {
	return &SettingsHandler{
		minioService: minioService,
		version:      version,
		commit:       commit,
		date:         date,
		builtBy:      builtBy,
	}
}

// ShowSettings displays the settings page
func (h *SettingsHandler) ShowSettings(c *gin.Context) {
	RenderWithTranslations(c, "settings.html", gin.H{
		"title":      "settings.title",
		"version":    h.version,
		"commit":     h.commit,
		"build_date": h.date,
		"built_by":   h.builtBy,
	})
}
