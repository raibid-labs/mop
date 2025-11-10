package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger creates a structured logging middleware
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request ID if available
		requestID, _ := c.Get(RequestIDKey)

		// Log request details
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if requestID != nil {
			fields = append(fields, zap.String("request_id", requestID.(string)))
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		// Choose log level based on status code
		if c.Writer.Status() >= 500 {
			logger.Error("http request", fields...)
		} else if c.Writer.Status() >= 400 {
			logger.Warn("http request", fields...)
		} else {
			logger.Info("http request", fields...)
		}
	}
}
