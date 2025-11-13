package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raibid-labs/mop/examples/04-redis-cache/internal/cache"
	"github.com/raibid-labs/mop/examples/04-redis-cache/internal/handlers"
)

func main() {
	// Configuration from environment
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	serverPort := getEnv("SERVER_PORT", "8080")

	// Initialize cache
	cacheConfig := cache.Config{
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		RedisDB:       0,
	}

	c, err := cache.New(cacheConfig)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Close()

	log.Printf("Connected to Redis at %s", redisAddr)

	// Initialize pub/sub
	ps := cache.NewPubSub(c)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ps.Subscribe(ctx); err != nil {
		log.Fatalf("Failed to subscribe to pub/sub: %v", err)
	}
	defer ps.Stop()

	// Initialize handlers
	h := handlers.New(c, ps)

	// Setup router
	router := gin.Default()

	// Health check
	router.GET("/health", h.Health)

	// Cache statistics
	router.GET("/stats", h.GetStats)
	router.POST("/stats/reset", h.ResetStats)

	// Item endpoints with caching
	router.GET("/items/:id", h.GetItem)
	router.POST("/items/batch", h.GetMultipleItems)
	router.POST("/items", h.CreateItem)
	router.PUT("/items/:id", h.UpdateItem)
	router.DELETE("/items/:id", h.DeleteItem)

	// Cache management
	router.POST("/cache/invalidate", h.InvalidateCache)

	// Start server
	go func() {
		log.Printf("Starting server on :%s", serverPort)
		if err := router.Run(":" + serverPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Cleanup
	cancel()
	time.Sleep(1 * time.Second)

	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
