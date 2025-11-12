package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raibid-labs/mop/examples/03-sql-app/internal/db"
	"github.com/raibid-labs/mop/examples/03-sql-app/internal/handlers"
	"github.com/raibid-labs/mop/examples/03-sql-app/internal/repository"
)

func main() {
	// Get configuration from environment
	dbConfig := db.Config{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnvAsInt("DB_PORT", 5432),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "postgres"),
		Database:        getEnv("DB_NAME", "orders"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxConns:        int32(getEnvAsInt("DB_MAX_CONNS", 25)),
		MinConns:        int32(getEnvAsInt("DB_MIN_CONNS", 5)),
		MaxConnLifetime: time.Duration(getEnvAsInt("DB_MAX_CONN_LIFETIME", 3600)) * time.Second,
		MaxConnIdleTime: time.Duration(getEnvAsInt("DB_MAX_CONN_IDLE_TIME", 300)) * time.Second,
	}

	serverPort := getEnv("SERVER_PORT", "8080")

	// Create database connection pool
	ctx := context.Background()
	pool, err := db.NewPool(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Successfully connected to database")

	// Initialize repositories
	customerRepo := repository.NewCustomerRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(pool)
	customerHandler := handlers.NewCustomerHandler(customerRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo)

	// Set up Gin router
	router := gin.Default()

	// Health check endpoints
	router.GET("/health", healthHandler.Check)
	router.GET("/ready", healthHandler.Ready)

	// Customer endpoints
	router.POST("/customers", customerHandler.Create)
	router.GET("/customers/:id", customerHandler.GetByID)
	router.GET("/customers", customerHandler.List)

	// Order endpoints
	router.POST("/orders", orderHandler.Create)
	router.GET("/orders/:id", orderHandler.GetByID)
	router.PUT("/orders/:id/status", orderHandler.UpdateStatus)
	router.GET("/customers/:customer_id/orders", orderHandler.ListByCustomer)
	router.GET("/customers/:customer_id/orders/stats", orderHandler.GetStats)

	// Slow query endpoint for OBI testing
	router.GET("/customers/:customer_id/orders/slow", orderHandler.SimulateSlowQuery)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + serverPort,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %s", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
