package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that allows all origins (safe for LAN-only deployment).
func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Range", "Content-Length"}
	config.ExposeHeaders = []string{"Content-Length", "Content-Range", "Accept-Ranges", "Content-Disposition"}
	config.AllowCredentials = false
	return cors.New(config)
}
