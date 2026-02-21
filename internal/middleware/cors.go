package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a GIN middleware that configures CORS headers.
// In development, all origins are allowed; in production, restrict to specific domains.
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}
