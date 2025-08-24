package middleware

import (
	"time"

	"github.com/feature-voting-platform/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Log request start
		utils.LogInfo("Request started",
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithMetadata("user_agent", c.GetHeader("User-Agent")),
			utils.WithMetadata("remote_addr", c.ClientIP()))

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Get user ID if available
		logFields := []utils.LogField{
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(c.Writer.Status()),
			utils.WithMetadata("latency_ms", latency.Milliseconds()),
			utils.WithMetadata("response_size", c.Writer.Size()),
		}

		if userID, exists := c.Get("user_id"); exists {
			logFields = append(logFields, utils.WithUserID(userID.(int)))
		}

		// Log completion
		if c.Writer.Status() >= 400 {
			utils.LogWarning("Request completed with error status", logFields...)
		} else {
			utils.LogInfo("Request completed successfully", logFields...)
		}
	}
}