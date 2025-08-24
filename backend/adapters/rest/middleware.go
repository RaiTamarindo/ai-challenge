package rest

import (
	"net/http"
	"strings"
	"time"

	"github.com/feature-voting-platform/backend/adapters/auth"
	"github.com/feature-voting-platform/backend/adapters/logs"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware returns a logging middleware
func LoggingMiddleware(logger logs.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Log request start
		logger.Info("Request started",
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithMetadata("user_agent", c.GetHeader("User-Agent")),
			logs.WithMetadata("remote_addr", c.ClientIP()))

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Get user ID if available
		logFields := []logs.LogField{
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(c.Writer.Status()),
			logs.WithMetadata("latency_ms", latency.Milliseconds()),
			logs.WithMetadata("response_size", c.Writer.Size()),
		}

		if userID, exists := c.Get("user_id"); exists {
			logFields = append(logFields, logs.WithUserID(userID.(int)))
		}

		// Log completion
		if c.Writer.Status() >= 400 {
			logger.Warning("Request completed with error status", logFields...)
		} else {
			logger.Info("Request completed successfully", logFields...)
		}
	}
}

// AuthMiddleware returns an authentication middleware
func AuthMiddleware(tokenService auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := tokenService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// OptionalAuthMiddleware returns an optional authentication middleware
func OptionalAuthMiddleware(tokenService auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := tokenService.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

		c.Next()
	}
}