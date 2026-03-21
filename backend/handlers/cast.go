package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"tea-cloud/backend/services"
)

// CastHandler exposes DLNA cast discovery and control endpoints.
type CastHandler struct {
	castSvc *services.CastService
	port    int
}

func NewCastHandler(castSvc *services.CastService, port int) *CastHandler {
	return &CastHandler{castSvc: castSvc, port: port}
}

// GET /api/cast/devices
// Discovers DLNA MediaRenderer devices on the LAN.
func (h *CastHandler) ListDevices(c *gin.Context) {
	renderers, err := h.castSvc.DiscoverRenderers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if renderers == nil {
		renderers = []services.DLNARenderer{}
	}
	c.JSON(http.StatusOK, renderers)
}

// POST /api/cast
// Body: { "location": "<renderer device URL>", "path": "<file path>" }
func (h *CastHandler) Cast(c *gin.Context) {
	var req struct {
		Location string `json:"location" binding:"required"`
		Path     string `json:"path"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mediaURL := h.castSvc.BuildMediaURL(req.Location, req.Path, h.port)
	title := services.TitleFromPath(req.Path)

	if err := h.castSvc.CastTo(req.Location, mediaURL, title); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "mediaURL": mediaURL})
}
