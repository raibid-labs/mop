package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimit creates a rate limiting middleware
func RateLimit(rps int) gin.HandlerFunc {
	// Store limiters per IP address
	type client struct {
		limiter  *rate.Limiter
		lastSeen int64
	}

	var (
		mu      sync.RWMutex
		clients = make(map[string]*client)
	)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		if _, exists := clients[ip]; !exists {
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(rps), rps*2),
			}
		}
		limiter := clients[ip].limiter
		mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"message": "Too many requests from this IP address",
			})
			return
		}

		c.Next()
	}
}
