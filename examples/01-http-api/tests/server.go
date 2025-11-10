package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/raibid-labs/mop/examples/01-http-api/internal/handlers"
	"github.com/raibid-labs/mop/examples/01-http-api/internal/middleware"
	"github.com/raibid-labs/mop/examples/01-http-api/internal/store"
)

func startTestServer(t *testing.T) *http.Server {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize store
	productStore := store.NewMemoryStore()

	// Create Gin router
	gin.SetMode(gin.TestMode)
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
		Addr:    ":18080",
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	return srv
}

func shutdownServer(ctx context.Context, srv *http.Server) error {
	return srv.Shutdown(ctx)
}

func makeRequest(method, url string, body []byte) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return client.Do(req)
}
