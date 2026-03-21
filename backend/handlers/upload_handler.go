package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"tea-cloud/backend/services"
)

// UploadHandler handles resumable chunked file uploads.
type UploadHandler struct {
	svc *services.UploadService
}

func NewUploadHandler(svc *services.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

type initUploadRequest struct {
	SaveDir     string `json:"saveDir"`
	Filename    string `json:"filename"`
	TotalSize   int64  `json:"totalSize"`
	TotalChunks int    `json:"totalChunks"`
}

// InitUpload godoc
// POST /api/uploads
// Creates a new upload session and returns uploadId + list of already-uploaded chunks.
func (h *UploadHandler) InitUpload(c *gin.Context) {
	var req initUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Filename == "" || req.TotalChunks <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename and totalChunks are required"})
		return
	}
	if req.SaveDir == "" {
		req.SaveDir = "."
	}

	session, err := h.svc.Init(req.Filename, req.SaveDir, req.TotalSize, req.TotalChunks)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, chunks, _ := h.svc.GetStatus(session.ID)
	if chunks == nil {
		chunks = []int{}
	}
	c.JSON(http.StatusOK, gin.H{
		"uploadId":       session.ID,
		"uploadedChunks": chunks,
	})
}

// GetStatus godoc
// GET /api/uploads/:id
// Returns the list of chunk indices already uploaded for the given session.
func (h *UploadHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")
	_, chunks, err := h.svc.GetStatus(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if chunks == nil {
		chunks = []int{}
	}
	c.JSON(http.StatusOK, gin.H{
		"uploadId":       id,
		"uploadedChunks": chunks,
	})
}

// UploadChunk godoc
// PUT /api/uploads/:id/chunk?index=N
// Receives raw bytes for a single chunk.
func (h *UploadHandler) UploadChunk(c *gin.Context) {
	id := c.Param("id")
	indexStr := c.Query("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chunk index"})
		return
	}

	if err := h.svc.SaveChunk(id, index, c.Request.Body); err != nil {
		if err.Error() == "upload session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// CompleteUpload godoc
// POST /api/uploads/:id/complete
// Merges all chunks into the final file and returns FileInfo.
func (h *UploadHandler) CompleteUpload(c *gin.Context) {
	id := c.Param("id")
	info, err := h.svc.Complete(id)
	if err != nil {
		if err.Error() == "upload session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}
