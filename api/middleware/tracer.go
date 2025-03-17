package middleware

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RetreiveIP returns middleware that extracts and converts the client's IP address
func RetreiveIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		IP := c.ClientIP()
		parsedIP := net.ParseIP(IP)
		if parsedIP != nil {
			if ipv4 := parsedIP.To4(); ipv4 != nil {
				IP = ipv4.String()
			}
		}
		c.Set("clientIP", IP)
		c.Next()
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// add security and CORS headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CORS headers
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		// Removing these headers because this is a pure API server
		// c.Header("X-Frame-Options", "DENY")
		// c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("X-XSS-Protection", "1; mode=block")

		// Handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept, Authorization")
			// Cache preflight response for 24 hours
			c.Header("Access-Control-Max-Age", "86400") // Cache preflight response for 24 hours
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
