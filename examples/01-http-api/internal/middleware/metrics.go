package middleware

import (
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	requestCount   uint64
	requestLatency uint64
)

// Metrics collects basic request metrics
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment request counter
		atomic.AddUint64(&requestCount, 1)

		c.Next()

		// Record latency in milliseconds
		latency := uint64(time.Since(start).Milliseconds())
		atomic.AddUint64(&requestLatency, latency)
	}
}

// GetRequestCount returns the total number of requests
func GetRequestCount() uint64 {
	return atomic.LoadUint64(&requestCount)
}

// GetTotalLatency returns the total latency of all requests
func GetTotalLatency() uint64 {
	return atomic.LoadUint64(&requestLatency)
}
