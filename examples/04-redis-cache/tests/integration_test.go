package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raibid-labs/mop/examples/04-redis-cache/internal/cache"
	"github.com/raibid-labs/mop/examples/04-redis-cache/internal/handlers"
	"github.com/raibid-labs/mop/examples/04-redis-cache/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a test server with cache and handlers
func setupTestServer(t *testing.T) (*gin.Engine, *cache.Cache, *cache.PubSub) {
	t.Helper()

	// Create cache
	cfg := cache.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       15, // Use test DB
	}

	c, err := cache.New(cfg)
	require.NoError(t, err, "Failed to create test cache")

	// Flush test database
	ctx := context.Background()
	err = c.Flush(ctx)
	require.NoError(t, err, "Failed to flush test database")

	// Create pub/sub
	ps := cache.NewPubSub(c)
	err = ps.Subscribe(context.Background())
	require.NoError(t, err, "Failed to subscribe to pub/sub")

	// Create handlers
	h := handlers.New(c, ps)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/health", h.Health)
	router.GET("/stats", h.GetStats)
	router.POST("/stats/reset", h.ResetStats)
	router.GET("/items/:id", h.GetItem)
	router.POST("/items/batch", h.GetMultipleItems)
	router.POST("/items", h.CreateItem)
	router.PUT("/items/:id", h.UpdateItem)
	router.DELETE("/items/:id", h.DeleteItem)
	router.POST("/cache/invalidate", h.InvalidateCache)

	return router, c, ps
}

func TestIntegration_Health(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "connected", response["redis"])
}

func TestIntegration_GetItemCacheMiss(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	// Reset stats
	c.ResetStats()

	// Request item that exists in mock data
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "MISS", w.Header().Get("X-Cache-Status"))

	var item models.Item
	err := json.Unmarshal(w.Body.Bytes(), &item)
	require.NoError(t, err)

	assert.Equal(t, "1", item.ID)
	assert.Equal(t, "Laptop Pro", item.Name)
}

func TestIntegration_GetItemCacheHit(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	// First request (cache miss)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/items/1", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, "MISS", w1.Header().Get("X-Cache-Status"))

	// Second request (cache hit)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/items/1", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, "HIT", w2.Header().Get("X-Cache-Status"))

	var item models.Item
	err := json.Unmarshal(w2.Body.Bytes(), &item)
	require.NoError(t, err)
	assert.Equal(t, "Laptop Pro", item.Name)
}

func TestIntegration_GetItemNotFound(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items/999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "MISS", w.Header().Get("X-Cache-Status"))
}

func TestIntegration_CreateItem(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	newItem := models.Item{
		ID:          "100",
		Name:        "New Product",
		Description: "A new test product",
		Price:       199.99,
		Category:    "Test",
		Stock:       5,
	}

	body, _ := json.Marshal(newItem)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var created models.Item
	err := json.Unmarshal(w.Body.Bytes(), &created)
	require.NoError(t, err)
	assert.Equal(t, "New Product", created.Name)
	assert.NotZero(t, created.CreatedAt)

	// Verify it's cached
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/items/100", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, "HIT", w2.Header().Get("X-Cache-Status"))
}

func TestIntegration_UpdateItem(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	// Create initial item
	initial := models.Item{
		ID:    "200",
		Name:  "Original Name",
		Price: 99.99,
	}
	body1, _ := json.Marshal(initial)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/items", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	// Update item
	updated := models.Item{
		Name:  "Updated Name",
		Price: 149.99,
	}
	body2, _ := json.Marshal(updated)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("PUT", "/items/200", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var result models.Item
	err := json.Unmarshal(w2.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", result.Name)
	assert.Equal(t, 149.99, result.Price)

	// Verify cache updated
	time.Sleep(100 * time.Millisecond) // Allow pub/sub to process

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/items/200", nil)
	router.ServeHTTP(w3, req3)

	var cached models.Item
	err = json.Unmarshal(w3.Body.Bytes(), &cached)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", cached.Name)
}

func TestIntegration_DeleteItem(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	// Create item
	item := models.Item{
		ID:   "300",
		Name: "To Be Deleted",
	}
	body, _ := json.Marshal(item)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/items", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	// Delete item
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("DELETE", "/items/300", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusNoContent, w2.Code)

	// Verify cache cleared
	time.Sleep(100 * time.Millisecond) // Allow pub/sub to process

	ctx := context.Background()
	exists, err := c.Exists(ctx, "item:300")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestIntegration_GetMultipleItems(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	// Request multiple items
	requestBody := map[string][]string{
		"ids": {"1", "2", "3"},
	}
	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/items/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	items := response["items"].([]interface{})
	assert.Equal(t, 3, len(items))

	// Second request should have all hits
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/items/batch", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	assert.Equal(t, "3", w2.Header().Get("X-Cache-Hits"))
	assert.Equal(t, "0", w2.Header().Get("X-Cache-Misses"))
}

func TestIntegration_Stats(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	// Reset stats
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/stats/reset", nil)
	router.ServeHTTP(w1, req1)

	// Generate some cache activity
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/items/1", nil) // Miss
	router.ServeHTTP(w2, req2)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/items/1", nil) // Hit
	router.ServeHTTP(w3, req3)

	// Get stats
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/stats", nil)
	router.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusOK, w4.Code)

	var stats models.CacheStats
	err := json.Unmarshal(w4.Body.Bytes(), &stats)
	require.NoError(t, err)

	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.InDelta(t, 0.5, stats.HitRate, 0.01)
}

func TestIntegration_CacheInvalidation(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	// Create item
	item := models.Item{
		ID:   "400",
		Name: "Invalidate Me",
	}
	body1, _ := json.Marshal(item)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/items", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	// Verify cached
	ctx := context.Background()
	exists, err := c.Exists(ctx, "item:400")
	require.NoError(t, err)
	assert.True(t, exists)

	// Invalidate
	invalidateReq := map[string]string{
		"key": "item:400",
	}
	body2, _ := json.Marshal(invalidateReq)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/cache/invalidate", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	// Wait for pub/sub processing
	time.Sleep(100 * time.Millisecond)

	// Verify invalidated
	exists, err = c.Exists(ctx, "item:400")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestIntegration_CacheInvalidationPattern(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	ctx := context.Background()

	// Create multiple items
	for i := 1; i <= 3; i++ {
		item := models.Item{
			ID:   "500" + string(rune('0'+i)),
			Name: "Pattern Test",
		}
		body, _ := json.Marshal(item)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/items", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}

	// Invalidate by pattern
	invalidateReq := map[string]string{
		"pattern": "item:500*",
	}
	body, _ := json.Marshal(invalidateReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cache/invalidate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Wait for pub/sub processing
	time.Sleep(100 * time.Millisecond)

	// Verify all invalidated
	for i := 1; i <= 3; i++ {
		key := "item:500" + string(rune('0'+i))
		exists, err := c.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists, "Key %s should be invalidated", key)
	}
}

func TestIntegration_InvalidRequest(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	// Invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/items", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	router, c, ps := setupTestServer(t)
	defer c.Close()
	defer ps.Stop()

	// Make concurrent requests
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/items/1", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			done <- true
		}()
	}

	// Wait for all requests
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify cache still works
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stats", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
