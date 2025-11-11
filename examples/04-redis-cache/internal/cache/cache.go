package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/raibid-labs/mop/examples/04-redis-cache/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	// DefaultTTL is the default cache expiration time
	DefaultTTL = 5 * time.Minute

	// InvalidationChannel is the pub/sub channel for cache invalidation
	InvalidationChannel = "cache:invalidate"
)

// Cache provides Redis caching with cache-aside pattern
type Cache struct {
	client *redis.Client
	hits   atomic.Int64
	misses atomic.Int64
}

// Config holds cache configuration
type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

// New creates a new Cache instance
func New(cfg Config) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &Cache{
		client: client,
	}, nil
}

// Get retrieves an item from cache
func (c *Cache) Get(ctx context.Context, key string) (*models.Item, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		c.misses.Add(1)
		return nil, nil // Cache miss, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	var item models.Item
	if err := json.Unmarshal([]byte(val), &item); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	c.hits.Add(1)
	return &item, nil
}

// Set stores an item in cache with TTL
func (c *Cache) Set(ctx context.Context, key string, item *models.Item, ttl time.Duration) error {
	if ttl == 0 {
		ttl = DefaultTTL
	}

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	return nil
}

// Delete removes an item from cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *Cache) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	pipe := c.client.Pipeline()

	for iter.Next(ctx) {
		pipe.Del(ctx, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("pipeline exec failed: %w", err)
	}

	return nil
}

// GetMultiple retrieves multiple items from cache using pipeline
func (c *Cache) GetMultiple(ctx context.Context, keys []string) (map[string]*models.Item, error) {
	if len(keys) == 0 {
		return make(map[string]*models.Item), nil
	}

	pipe := c.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		// Don't fail on Nil errors, as some keys might not exist
	}

	results := make(map[string]*models.Item)
	for key, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			c.misses.Add(1)
			continue // Skip missing keys
		}
		if err != nil {
			return nil, fmt.Errorf("get failed for key %s: %w", key, err)
		}

		var item models.Item
		if err := json.Unmarshal([]byte(val), &item); err != nil {
			return nil, fmt.Errorf("unmarshal failed for key %s: %w", key, err)
		}

		results[key] = &item
		c.hits.Add(1)
	}

	return results, nil
}

// SetMultiple stores multiple items in cache using pipeline
func (c *Cache) SetMultiple(ctx context.Context, items map[string]*models.Item, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	if ttl == 0 {
		ttl = DefaultTTL
	}

	pipe := c.client.Pipeline()

	for key, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("marshal failed for key %s: %w", key, err)
		}
		pipe.Set(ctx, key, data, ttl)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("pipeline exec failed: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return n > 0, nil
}

// GetTTL returns the remaining TTL for a key
func (c *Cache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl failed: %w", err)
	}
	return ttl, nil
}

// Expire sets a new TTL for a key
func (c *Cache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}
	return nil
}

// Stats returns cache performance statistics
func (c *Cache) Stats(ctx context.Context) (*models.CacheStats, error) {
	hits := c.hits.Load()
	misses := c.misses.Load()
	total := hits + misses

	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	// Count total keys using DBSIZE
	totalKeys, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis dbsize failed: %w", err)
	}

	return &models.CacheStats{
		Hits:      hits,
		Misses:    misses,
		HitRate:   hitRate,
		TotalKeys: totalKeys,
	}, nil
}

// ResetStats resets cache statistics
func (c *Cache) ResetStats() {
	c.hits.Store(0)
	c.misses.Store(0)
}

// Flush removes all keys from the cache (use with caution!)
func (c *Cache) Flush(ctx context.Context) error {
	if err := c.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("redis flushdb failed: %w", err)
	}
	return nil
}

// Close closes the Redis client connection
func (c *Cache) Close() error {
	return c.client.Close()
}

// Client returns the underlying Redis client (for pub/sub)
func (c *Cache) Client() *redis.Client {
	return c.client
}
