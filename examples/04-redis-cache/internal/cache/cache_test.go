package cache

import (
	"context"
	"testing"
	"time"

	"github.com/raibid-labs/mop/examples/04-redis-cache/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestCache creates a cache instance for testing
// Note: Requires a Redis instance running on localhost:6379
func getTestCache(t *testing.T) *Cache {
	t.Helper()

	cfg := Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       15, // Use DB 15 for testing
	}

	c, err := New(cfg)
	require.NoError(t, err, "Failed to create test cache")

	// Clean up before test
	ctx := context.Background()
	err = c.Flush(ctx)
	require.NoError(t, err, "Failed to flush test database")

	return c
}

func TestNew(t *testing.T) {
	cfg := Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       15,
	}

	c, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, c)

	defer c.Close()

	// Test connection
	ctx := context.Background()
	err = c.Client().Ping(ctx).Err()
	assert.NoError(t, err)
}

func TestCache_SetAndGet(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()
	key := "test:item:1"

	item := &models.Item{
		ID:          "1",
		Name:        "Test Item",
		Description: "Test Description",
		Price:       99.99,
		Category:    "Test",
		Stock:       10,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set item
	err := c.Set(ctx, key, item, 1*time.Minute)
	require.NoError(t, err)

	// Get item
	retrieved, err := c.Get(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, item.ID, retrieved.ID)
	assert.Equal(t, item.Name, retrieved.Name)
	assert.Equal(t, item.Price, retrieved.Price)
}

func TestCache_GetMiss(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()
	key := "test:nonexistent"

	item, err := c.Get(ctx, key)
	require.NoError(t, err)
	assert.Nil(t, item)

	// Verify miss counter incremented
	stats, err := c.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.Misses)
}

func TestCache_Delete(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()
	key := "test:item:delete"

	item := &models.Item{
		ID:   "delete",
		Name: "Delete Me",
	}

	// Set item
	err := c.Set(ctx, key, item, 1*time.Minute)
	require.NoError(t, err)

	// Verify exists
	exists, err := c.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete item
	err = c.Delete(ctx, key)
	require.NoError(t, err)

	// Verify deleted
	exists, err = c.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCache_DeletePattern(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()

	// Create multiple items
	for i := 1; i <= 3; i++ {
		key := "test:pattern:item:" + string(rune('0'+i))
		item := &models.Item{ID: string(rune('0' + i)), Name: "Item " + string(rune('0'+i))}
		err := c.Set(ctx, key, item, 1*time.Minute)
		require.NoError(t, err)
	}

	// Delete by pattern
	err := c.DeletePattern(ctx, "test:pattern:item:*")
	require.NoError(t, err)

	// Verify all deleted
	for i := 1; i <= 3; i++ {
		key := "test:pattern:item:" + string(rune('0'+i))
		exists, err := c.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestCache_GetMultiple(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()

	// Set multiple items
	items := map[string]*models.Item{
		"test:multi:1": {ID: "1", Name: "Item 1", Price: 10.0},
		"test:multi:2": {ID: "2", Name: "Item 2", Price: 20.0},
		"test:multi:3": {ID: "3", Name: "Item 3", Price: 30.0},
	}

	for key, item := range items {
		err := c.Set(ctx, key, item, 1*time.Minute)
		require.NoError(t, err)
	}

	// Get multiple
	keys := []string{"test:multi:1", "test:multi:2", "test:multi:3", "test:multi:missing"}
	results, err := c.GetMultiple(ctx, keys)
	require.NoError(t, err)

	assert.Equal(t, 3, len(results))
	assert.NotNil(t, results["test:multi:1"])
	assert.NotNil(t, results["test:multi:2"])
	assert.NotNil(t, results["test:multi:3"])
	assert.Nil(t, results["test:multi:missing"])

	assert.Equal(t, "Item 1", results["test:multi:1"].Name)
	assert.Equal(t, 20.0, results["test:multi:2"].Price)
}

func TestCache_SetMultiple(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()

	items := map[string]*models.Item{
		"test:batch:1": {ID: "1", Name: "Batch 1"},
		"test:batch:2": {ID: "2", Name: "Batch 2"},
		"test:batch:3": {ID: "3", Name: "Batch 3"},
	}

	// Set multiple
	err := c.SetMultiple(ctx, items, 1*time.Minute)
	require.NoError(t, err)

	// Verify all set
	for key, expected := range items {
		retrieved, err := c.Get(ctx, key)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, expected.Name, retrieved.Name)
	}
}

func TestCache_Exists(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()
	key := "test:exists"

	// Should not exist
	exists, err := c.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	// Create item
	item := &models.Item{ID: "1", Name: "Exists Test"}
	err = c.Set(ctx, key, item, 1*time.Minute)
	require.NoError(t, err)

	// Should exist now
	exists, err = c.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestCache_TTL(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()
	key := "test:ttl"

	item := &models.Item{ID: "1", Name: "TTL Test"}
	err := c.Set(ctx, key, item, 10*time.Second)
	require.NoError(t, err)

	// Get TTL
	ttl, err := c.GetTTL(ctx, key)
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, 10*time.Second)
}

func TestCache_Expire(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()
	key := "test:expire"

	item := &models.Item{ID: "1", Name: "Expire Test"}
	err := c.Set(ctx, key, item, 1*time.Minute)
	require.NoError(t, err)

	// Set new TTL
	err = c.Expire(ctx, key, 5*time.Second)
	require.NoError(t, err)

	// Verify new TTL
	ttl, err := c.GetTTL(ctx, key)
	require.NoError(t, err)
	assert.LessOrEqual(t, ttl, 5*time.Second)
}

func TestCache_Stats(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()

	// Reset stats
	c.ResetStats()

	// Generate some hits and misses
	key := "test:stats"
	item := &models.Item{ID: "1", Name: "Stats Test"}

	// Miss
	_, err := c.Get(ctx, key)
	require.NoError(t, err)

	// Set
	err = c.Set(ctx, key, item, 1*time.Minute)
	require.NoError(t, err)

	// Hit
	_, err = c.Get(ctx, key)
	require.NoError(t, err)

	// Hit
	_, err = c.Get(ctx, key)
	require.NoError(t, err)

	// Get stats
	stats, err := c.Stats(ctx)
	require.NoError(t, err)

	assert.Equal(t, int64(2), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.InDelta(t, 0.666, stats.HitRate, 0.01)
	assert.GreaterOrEqual(t, stats.TotalKeys, int64(1))
}

func TestCache_ResetStats(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()

	// Generate some activity
	key := "test:reset"
	_, _ = c.Get(ctx, key)

	stats, err := c.Stats(ctx)
	require.NoError(t, err)
	assert.Greater(t, stats.Misses, int64(0))

	// Reset
	c.ResetStats()

	stats, err = c.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
}

func TestCache_NewWithBadConfig(t *testing.T) {
	cfg := Config{
		RedisAddr:     "invalid:99999",
		RedisPassword: "",
		RedisDB:       0,
	}

	_, err := New(cfg)
	assert.Error(t, err)
}

func TestCache_InvalidJSON(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()
	key := "test:invalid"

	// Manually set invalid JSON
	err := c.Client().Set(ctx, key, "invalid json {]", 1*time.Minute).Err()
	require.NoError(t, err)

	// Try to get - should fail to unmarshal
	_, err = c.Get(ctx, key)
	assert.Error(t, err)
}

func TestCache_Pipeline(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()

	// Test empty arrays
	results, err := c.GetMultiple(ctx, []string{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(results))

	err = c.SetMultiple(ctx, map[string]*models.Item{}, 1*time.Minute)
	require.NoError(t, err)
}

// TestCache_ConcurrentAccess tests concurrent cache operations
func TestCache_ConcurrentAccess(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx := context.Background()
	key := "test:concurrent"

	// Multiple goroutines reading/writing
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			item := &models.Item{
				ID:   "concurrent",
				Name: "Concurrent Test",
			}
			_ = c.Set(ctx, key, item, 1*time.Minute)
			_, _ = c.Get(ctx, key)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify cache still works
	exists, err := c.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestCache_NilContext(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	// All operations should handle context properly
	ctx := context.Background()

	item := &models.Item{ID: "1", Name: "Context Test"}
	err := c.Set(ctx, "test:ctx", item, 1*time.Minute)
	assert.NoError(t, err)
}

func TestCache_Timeout(t *testing.T) {
	c := getTestCache(t)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Normal operation should work
	item := &models.Item{ID: "1", Name: "Timeout Test"}
	err := c.Set(ctx, "test:timeout", item, 1*time.Minute)
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkCache_Set(b *testing.B) {
	c := getTestCache(&testing.T{})
	defer c.Close()

	ctx := context.Background()
	item := &models.Item{ID: "bench", Name: "Benchmark Test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Set(ctx, "bench:set", item, 1*time.Minute)
	}
}

func BenchmarkCache_Get(b *testing.B) {
	c := getTestCache(&testing.T{})
	defer c.Close()

	ctx := context.Background()
	item := &models.Item{ID: "bench", Name: "Benchmark Test"}
	_ = c.Set(ctx, "bench:get", item, 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Get(ctx, "bench:get")
	}
}

func BenchmarkCache_GetMultiple(b *testing.B) {
	c := getTestCache(&testing.T{})
	defer c.Close()

	ctx := context.Background()
	keys := []string{"bench:1", "bench:2", "bench:3", "bench:4", "bench:5"}

	for _, key := range keys {
		item := &models.Item{ID: key, Name: "Benchmark"}
		_ = c.Set(ctx, key, item, 1*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.GetMultiple(ctx, keys)
	}
}
