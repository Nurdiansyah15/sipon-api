package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Untuk production, ganti dengan domain spesifik
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-Id", "X-Client-Request-Id"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
