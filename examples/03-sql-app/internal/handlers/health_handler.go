package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db *pgxpool.Pool
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check performs a health check
func (h *HealthHandler) Check(c *gin.Context) {
	// Check database connection
	if err := h.db.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unhealthy",
			"database": "disconnected",
			"error":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"database": "connected",
	})
}

// Ready performs a readiness check
func (h *HealthHandler) Ready(c *gin.Context) {
	// Check database connection
	if err := h.db.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready":    false,
			"database": "not ready",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ready":    true,
		"database": "ready",
	})
}
