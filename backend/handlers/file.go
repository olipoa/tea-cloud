package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"tea-cloud/backend/services"
)

type FileHandler struct {
	svc *services.FileService
}

func NewFileHandler(svc *services.FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

// ListFiles godoc
// GET /api/files?path=
// Returns a JSON array of FileInfo for the given directory path (default: root).
func (h *FileHandler) ListFiles(c *gin.Context) {
	relPath := c.DefaultQuery("path", ".")
	items, err := h.svc.ListDir(relPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "directory not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetFileInfo godoc
// GET /api/files/info?path=
func (h *FileHandler) GetFileInfo(c *gin.Context) {
	relPath := c.Query("path")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}
	info, err := h.svc.GetFileInfo(relPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

// DownloadFile godoc
// GET /api/files/download?path=
// Supports HTTP Range requests for video/audio streaming.
func (h *FileHandler) DownloadFile(c *gin.Context) {
	relPath := c.Query("path")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}

	f, info, err := h.svc.OpenFile(relPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	modTime := time.UnixMilli(info.ModTime)

	// Let net/http handle Range, ETag, Last-Modified, and Content-Type
	if info.MIME != "" {
		c.Header("Content-Type", info.MIME)
	}
	if c.Query("download") == "1" {
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(info.Name)))
	} else {
		c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filepath.Base(info.Name)))
	}
	c.Header("Accept-Ranges", "bytes")
	http.ServeContent(c.Writer, c.Request, info.Name, modTime, f)
}

// UploadFile godoc
// POST /api/files/upload?path=  (multipart form, field "file")
func (h *FileHandler) UploadFile(c *gin.Context) {
	relDir := c.DefaultQuery("path", ".")

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}

	var uploaded []services.FileInfo
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open uploaded file"})
			return
		}
		info, err := h.svc.SaveFile(relDir, fh.Filename, src)
		src.Close()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		uploaded = append(uploaded, *info)
	}

	c.JSON(http.StatusOK, gin.H{"uploaded": uploaded})
}

// DeleteFile godoc
// DELETE /api/files?path=
func (h *FileHandler) DeleteFile(c *gin.Context) {
	relPath := c.Query("path")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}
	if err := h.svc.DeletePath(relPath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": relPath})
}

// CreateDir godoc
// POST /api/dirs?path=
func (h *FileHandler) CreateDir(c *gin.Context) {
	relPath := c.Query("path")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}
	if err := h.svc.MkDir(relPath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"created": relPath})
}

// StaticFileMiddleware serves raw files (used for inline preview).
// GET /raw/*filepath
func StaticFileMiddleware(svc *services.FileService) gin.HandlerFunc {
	return func(c *gin.Context) {
		relPath := c.Param("filepath")
		if relPath == "" || relPath == "/" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
			return
		}
		// Strip leading slash
		if len(relPath) > 0 && relPath[0] == '/' {
			relPath = relPath[1:]
		}

		log.Printf("[raw] %s %s -> %q", c.Request.Method, c.Request.RemoteAddr, relPath)

		abs, err := svc.AbsPath(relPath)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		f, err := os.Open(abs)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "not a file"})
			return
		}

		info, _ := svc.GetFileInfo(relPath)
		if info != nil && info.MIME != "" {
			c.Header("Content-Type", info.MIME)
		}
		c.Header("Accept-Ranges", "bytes")
		c.Header("Content-Length", strconv.FormatInt(stat.Size(), 10))
		http.ServeContent(c.Writer, c.Request, stat.Name(), stat.ModTime(), f)
	}
}
