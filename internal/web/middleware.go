package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logging middleware
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		timeStamp := time.Now()
		latency := timeStamp.Sub(start)

		// Get request info
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		// Log request
		logrus.WithFields(logrus.Fields{
			"status":    statusCode,
			"latency":   latency,
			"client_ip": clientIP,
			"method":    method,
			"path":      path,
			"body_size": bodySize,
		}).Info("HTTP Request")
	}
}

// Rate limiting middleware (simple implementation)
func RateLimitMiddleware() gin.HandlerFunc {
	type client struct {
		requests int
		lastSeen time.Time
	}

	clients := make(map[string]*client)
	const maxRequests = 100
	const windowSize = time.Minute

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		// Clean up old entries
		for k, v := range clients {
			if now.Sub(v.lastSeen) > windowSize {
				delete(clients, k)
			}
		}

		// Check current client
		if cl, exists := clients[ip]; exists {
			if now.Sub(cl.lastSeen) > windowSize {
				cl.requests = 1
				cl.lastSeen = now
			} else {
				cl.requests++
				if cl.requests > maxRequests {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "Rate limit exceeded",
					})
					c.Abort()
					return
				}
			}
		} else {
			clients[ip] = &client{
				requests: 1,
				lastSeen: now,
			}
		}

		c.Next()
	}
}

// Authentication middleware
func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for certain paths
		skipPaths := []string{
			"/api/auth/",
			"/static/",
			"/",
			"/login",
			"/ws",
		}

		for _, path := range skipPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// Check if user is authenticated
		if !s.twitchClient.IsLoggedIn() {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Error handling middleware
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// Security headers middleware
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; style-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; img-src 'self' data: https:; connect-src 'self' wss: https:;")

		c.Next()
	}
}
