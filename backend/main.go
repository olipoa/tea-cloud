package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"tea-cloud/backend/config"
	"tea-cloud/backend/handlers"
	"tea-cloud/backend/middleware"
	"tea-cloud/backend/services"
)

//go:embed static
var staticFiles embed.FS

func main() {
	cfg := config.Load()

	log.Printf("===========================================")
	log.Printf("  Tea Cloud - LAN File Sharing")
	log.Printf("===========================================")
	log.Printf("  Share directory : %s", cfg.ShareDir)
	log.Printf("  Node name       : %s", cfg.NodeName)

	// Services
	fileSvc := services.NewFileService(cfg.ShareDir)
	discoverySvc := services.NewDiscoveryService(cfg.NodeName, cfg.Port)

	// Register mDNS (non-fatal if it fails)
	if err := discoverySvc.Register(); err != nil {
		log.Printf("[WARN] mDNS registration failed: %v", err)
	}
	defer discoverySvc.Stop()

	// Router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORS())

	// Increase multipart memory limit to 32 MB (larger files stream to disk)
	r.MaxMultipartMemory = 32 << 20

	// API routes
	api := r.Group("/api")
	{
		fileHandler := handlers.NewFileHandler(fileSvc)
		api.GET("/files", fileHandler.ListFiles)
		api.GET("/files/info", fileHandler.GetFileInfo)
		api.GET("/files/download", fileHandler.DownloadFile)
		api.POST("/files/upload", fileHandler.UploadFile)
		api.DELETE("/files", fileHandler.DeleteFile)
		api.POST("/dirs", fileHandler.CreateDir)

		thumbnailHandler := handlers.NewThumbnailHandler(services.NewThumbnailService(fileSvc))
		api.GET("/files/thumbnail", thumbnailHandler.GetThumbnail)

		uploadHandler := handlers.NewUploadHandler(services.NewUploadService(fileSvc))
		api.POST("/uploads", uploadHandler.InitUpload)
		api.GET("/uploads/:id", uploadHandler.GetStatus)
		api.PUT("/uploads/:id/chunk", uploadHandler.UploadChunk)
		api.POST("/uploads/:id/complete", uploadHandler.CompleteUpload)

		peerHandler := handlers.NewPeerHandler(discoverySvc)
		api.GET("/peers", peerHandler.ListPeers)
		api.GET("/self", handlers.SelfInfo(cfg.NodeName, cfg.Port))

		castHandler := handlers.NewCastHandler(services.NewCastService(), cfg.Port)
		api.GET("/cast/devices", castHandler.ListDevices)
		api.POST("/cast", castHandler.Cast)
	}

	// Raw file serving (for inline preview / media streaming)
	r.GET("/raw/*filepath", handlers.StaticFileMiddleware(fileSvc))

	// Serve embedded Vue SPA
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("Failed to get static sub-filesystem: %v", err)
	}
	staticHandler := http.FileServer(http.FS(subFS))

	// All other routes → SPA (handles Vue Router history mode)
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Only attempt to serve a static asset if the path looks like a file
		// (has an extension) and actually exists in the embedded FS.
		// This avoids redirect loops caused by the file server trying to list dirs.
		if path != "/" && len(path) > 1 {
			rel := path[1:] // strip leading slash
			if info, err := fs.Stat(subFS, rel); err == nil && !info.IsDir() {
				staticHandler.ServeHTTP(c.Writer, c.Request)
				return
			}
		}
		// Fallback: serve index.html bytes directly (no redirects)
		data, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Print access URLs
	ips := services.GetLocalIPs()
	log.Printf("-------------------------------------------")
	log.Printf("  Access via browser:")
	log.Printf("  Local   : http://localhost:%d", cfg.Port)
	for _, ip := range ips {
		log.Printf("  Network : http://%s:%d", ip, cfg.Port)
	}
	log.Printf("===========================================")

	addr := fmt.Sprintf(":%d", cfg.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
