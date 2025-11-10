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
	"go.uber.org/zap"

	"github.com/raibid-labs/mop/examples/01-http-api/internal/handlers"
	"github.com/raibid-labs/mop/examples/01-http-api/internal/middleware"
	"github.com/raibid-labs/mop/examples/01-http-api/internal/store"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize store
	productStore := store.NewMemoryStore()

	// Create Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Apply middleware in order
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS())
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.RateLimit(100))
	r.Use(middleware.Metrics())

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productStore)
	healthHandler := handlers.NewHealthHandler()
	testingHandler := handlers.NewTestingHandler()

	// Register routes
	products := r.Group("/products")
	{
		products.GET("", productHandler.List)
		products.GET("/:id", productHandler.Get)
		products.POST("", productHandler.Create)
		products.PUT("/:id", productHandler.Update)
		products.DELETE("/:id", productHandler.Delete)
	}

	r.GET("/search", productHandler.Search)
	r.GET("/health", healthHandler.Health)
	r.GET("/slow", testingHandler.Slow)
	r.GET("/error", testingHandler.Error)

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
