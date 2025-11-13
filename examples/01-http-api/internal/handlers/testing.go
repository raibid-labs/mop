package handlers

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TestingHandler handles testing endpoints (slow, error)
type TestingHandler struct {
	rng *rand.Rand
}

// NewTestingHandler creates a new testing handler
func NewTestingHandler() *TestingHandler {
	return &TestingHandler{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Slow simulates a slow endpoint with 1-3 second latency
func (h *TestingHandler) Slow(c *gin.Context) {
	// Random delay between 1-3 seconds
	delay := time.Duration(1000+h.rng.Intn(2000)) * time.Millisecond
	time.Sleep(delay)

	c.JSON(http.StatusOK, gin.H{
		"message": "Slow endpoint response",
		"delay_ms": delay.Milliseconds(),
	})
}

// Error always returns a 500 internal server error
func (h *TestingHandler) Error(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "Simulated internal server error",
		"code":  "TESTING_ERROR",
	})
}
