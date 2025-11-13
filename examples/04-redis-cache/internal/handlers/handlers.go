package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raibid-labs/mop/examples/04-redis-cache/internal/cache"
	"github.com/raibid-labs/mop/examples/04-redis-cache/internal/models"
)

// Handler manages HTTP endpoints with caching
type Handler struct {
	cache  *cache.Cache
	pubsub *cache.PubSub
}

// New creates a new Handler
func New(c *cache.Cache, ps *cache.PubSub) *Handler {
	return &Handler{
		cache:  c,
		pubsub: ps,
	}
}

// GetItem retrieves an item by ID (with cache-aside pattern)
func (h *Handler) GetItem(c *gin.Context) {
	id := c.Param("id")
	cacheKey := fmt.Sprintf("item:%s", id)

	// Try to get from cache first
	item, err := h.cache.Get(c.Request.Context(), cacheKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cache error", "details": err.Error()})
		return
	}

	// Cache hit
	if item != nil {
		c.Header("X-Cache-Status", "HIT")
		c.JSON(http.StatusOK, item)
		return
	}

	// Cache miss - fetch from "upstream" (simulate with mock data)
	c.Header("X-Cache-Status", "MISS")
	item = h.fetchFromUpstream(id)
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	// Store in cache for next time
	if err := h.cache.Set(c.Request.Context(), cacheKey, item, cache.DefaultTTL); err != nil {
		// Log error but still return the item
		c.Header("X-Cache-Write-Error", err.Error())
	}

	c.JSON(http.StatusOK, item)
}

// GetMultipleItems retrieves multiple items by IDs (with pipeline)
func (h *Handler) GetMultipleItems(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Build cache keys
	keys := make([]string, len(req.IDs))
	for i, id := range req.IDs {
		keys[i] = fmt.Sprintf("item:%s", id)
	}

	// Get from cache using pipeline
	cached, err := h.cache.GetMultiple(c.Request.Context(), keys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cache error", "details": err.Error()})
		return
	}

	// Determine which items need to be fetched from upstream
	items := make([]*models.Item, 0, len(req.IDs))
	toCache := make(map[string]*models.Item)

	for _, id := range req.IDs {
		cacheKey := fmt.Sprintf("item:%s", id)
		if item, ok := cached[cacheKey]; ok {
			items = append(items, item)
		} else {
			// Fetch from upstream
			item := h.fetchFromUpstream(id)
			if item != nil {
				items = append(items, item)
				toCache[cacheKey] = item
			}
		}
	}

	// Cache the newly fetched items
	if len(toCache) > 0 {
		if err := h.cache.SetMultiple(c.Request.Context(), toCache, cache.DefaultTTL); err != nil {
			c.Header("X-Cache-Write-Error", err.Error())
		}
	}

	c.Header("X-Cache-Hits", fmt.Sprintf("%d", len(cached)))
	c.Header("X-Cache-Misses", fmt.Sprintf("%d", len(toCache)))
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CreateItem creates a new item and caches it
func (h *Handler) CreateItem(c *gin.Context) {
	var item models.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Set timestamps
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	// In a real app, save to database here
	// For demo, just cache it
	cacheKey := fmt.Sprintf("item:%s", item.ID)
	if err := h.cache.Set(c.Request.Context(), cacheKey, &item, cache.DefaultTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cache error", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdateItem updates an item and invalidates cache
func (h *Handler) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	var item models.Item

	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	item.ID = id
	item.UpdatedAt = time.Now()

	// Update in database (simulated)
	// Then update cache
	cacheKey := fmt.Sprintf("item:%s", id)
	if err := h.cache.Set(c.Request.Context(), cacheKey, &item, cache.DefaultTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cache error", "details": err.Error()})
		return
	}

	// Publish invalidation for other instances
	if err := h.pubsub.InvalidateKey(c.Request.Context(), cacheKey); err != nil {
		c.Header("X-Invalidation-Error", err.Error())
	}

	c.JSON(http.StatusOK, item)
}

// DeleteItem deletes an item and invalidates cache
func (h *Handler) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	cacheKey := fmt.Sprintf("item:%s", id)

	// Delete from database (simulated)
	// Then delete from cache
	if err := h.cache.Delete(c.Request.Context(), cacheKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cache error", "details": err.Error()})
		return
	}

	// Publish invalidation for other instances
	if err := h.pubsub.InvalidateKey(c.Request.Context(), cacheKey); err != nil {
		c.Header("X-Invalidation-Error", err.Error())
	}

	c.JSON(http.StatusNoContent, nil)
}

// InvalidateCache manually invalidates cache by key or pattern
func (h *Handler) InvalidateCache(c *gin.Context) {
	var req struct {
		Key     string `json:"key"`
		Pattern string `json:"pattern"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	if req.Key == "" && req.Pattern == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key or pattern required"})
		return
	}

	if req.Key != "" {
		if err := h.pubsub.InvalidateKey(c.Request.Context(), req.Key); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalidation failed", "details": err.Error()})
			return
		}
	}

	if req.Pattern != "" {
		if err := h.pubsub.InvalidatePattern(c.Request.Context(), req.Pattern); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalidation failed", "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "invalidation published"})
}

// GetStats returns cache statistics
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.cache.Stats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "stats error", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ResetStats resets cache statistics
func (h *Handler) ResetStats(c *gin.Context) {
	h.cache.ResetStats()
	c.JSON(http.StatusOK, gin.H{"message": "stats reset"})
}

// Health returns health status
func (h *Handler) Health(c *gin.Context) {
	// Check Redis connection
	if err := h.cache.Client().Ping(c.Request.Context()).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"redis":  "disconnected",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"redis":  "connected",
	})
}

// fetchFromUpstream simulates fetching data from an upstream service
// In a real application, this would call a database or external API
func (h *Handler) fetchFromUpstream(id string) *models.Item {
	// Simulate network delay
	time.Sleep(50 * time.Millisecond)

	// Mock data - in production, fetch from database
	mockItems := map[string]*models.Item{
		"1": {
			ID:          "1",
			Name:        "Laptop Pro",
			Description: "High-performance laptop for developers",
			Price:       1299.99,
			Category:    "Electronics",
			Stock:       15,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
		"2": {
			ID:          "2",
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse",
			Price:       29.99,
			Category:    "Accessories",
			Stock:       50,
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
		"3": {
			ID:          "3",
			Name:        "Mechanical Keyboard",
			Description: "RGB mechanical gaming keyboard",
			Price:       149.99,
			Category:    "Accessories",
			Stock:       25,
			CreatedAt:   time.Now().Add(-72 * time.Hour),
			UpdatedAt:   time.Now().Add(-3 * time.Hour),
		},
	}

	return mockItems[id]
}
