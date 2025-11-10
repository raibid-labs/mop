package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/raibid-labs/mop/examples/01-http-api/internal/handlers"
	"github.com/raibid-labs/mop/examples/01-http-api/internal/middleware"
	"github.com/raibid-labs/mop/examples/01-http-api/internal/store"
)

func setupBenchmarkRouter() *gin.Engine {
	logger, _ := zap.NewProduction()
	productStore := store.NewMemoryStore()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS())
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.RateLimit(10000)) // High limit for benchmarks
	r.Use(middleware.Metrics())

	productHandler := handlers.NewProductHandler(productStore)
	healthHandler := handlers.NewHealthHandler()

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

	// Pre-populate with some data
	for i := 0; i < 100; i++ {
		productStore.Create(&struct {
			ID          string
			Name        string
			Description string
			Price       float64
			Stock       int
			CreatedAt   time.Time
			UpdatedAt   time.Time
		}{
			Name:  "Benchmark Product",
			Price: 99.99,
			Stock: 100,
		})
	}

	return r
}

func BenchmarkHealthCheck(b *testing.B) {
	r := setupBenchmarkRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkListProducts(b *testing.B) {
	r := setupBenchmarkRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkCreateProduct(b *testing.B) {
	r := setupBenchmarkRouter()

	product := map[string]interface{}{
		"name":  "Benchmark Product",
		"price": 99.99,
		"stock": 100,
	}
	body, _ := json.Marshal(product)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkSearch(b *testing.B) {
	r := setupBenchmarkRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/search?q=Benchmark", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkConcurrentRequests(b *testing.B) {
	r := setupBenchmarkRouter()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}
