package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raibid-labs/mop/examples/03-sql-app/internal/db"
	"github.com/raibid-labs/mop/examples/03-sql-app/internal/handlers"
	"github.com/raibid-labs/mop/examples/03-sql-app/internal/models"
	"github.com/raibid-labs/mop/examples/03-sql-app/internal/repository"
)

var (
	testDB     *pgxpool.Pool
	testRouter *gin.Engine
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	// Use test database configuration
	cfg := db.Config{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		Database: "orders_test",
		SSLMode:  "disable",
		MaxConns: 10,
		MinConns: 2,
	}

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up database
	cleanupDB(t, pool)

	// Run migrations
	runMigrations(t, pool)

	return pool
}

func cleanupDB(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	queries := []string{
		"DROP TABLE IF EXISTS order_items CASCADE",
		"DROP TABLE IF EXISTS orders CASCADE",
		"DROP TABLE IF EXISTS customers CASCADE",
		"DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE",
	}

	for _, query := range queries {
		if _, err := pool.Exec(ctx, query); err != nil {
			t.Logf("Warning: cleanup query failed: %v", err)
		}
	}
}

func runMigrations(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Create customers table
	customersSQL := `
		CREATE TABLE IF NOT EXISTS customers (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
	`

	if _, err := pool.Exec(ctx, customersSQL); err != nil {
		t.Fatalf("Failed to create customers table: %v", err)
	}

	// Create orders and order_items tables
	ordersSQL := `
		CREATE TABLE IF NOT EXISTS orders (
			id BIGSERIAL PRIMARY KEY,
			customer_id BIGINT NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			total DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_orders_customer FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
		CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
		CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at DESC);

		CREATE TABLE IF NOT EXISTS order_items (
			id BIGSERIAL PRIMARY KEY,
			order_id BIGINT NOT NULL,
			product_id BIGINT NOT NULL,
			quantity INT NOT NULL CHECK (quantity > 0),
			price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
		CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);

		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
		CREATE TRIGGER update_orders_updated_at
			BEFORE UPDATE ON orders
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
	`

	if _, err := pool.Exec(ctx, ordersSQL); err != nil {
		t.Fatalf("Failed to create orders tables: %v", err)
	}
}

func setupRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize repositories
	customerRepo := repository.NewCustomerRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(pool)
	customerHandler := handlers.NewCustomerHandler(customerRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo)

	// Routes
	router.GET("/health", healthHandler.Check)
	router.POST("/customers", customerHandler.Create)
	router.GET("/customers/:id", customerHandler.GetByID)
	router.GET("/customers", customerHandler.List)
	router.POST("/orders", orderHandler.Create)
	router.GET("/orders/:id", orderHandler.GetByID)
	router.PUT("/orders/:id/status", orderHandler.UpdateStatus)
	router.GET("/customers/:customer_id/orders", orderHandler.ListByCustomer)
	router.GET("/customers/:customer_id/orders/stats", orderHandler.GetStats)

	return router
}

func TestHealthCheck(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := setupRouter(pool)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
}

func TestCreateCustomer(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := setupRouter(pool)

	customer := map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	body, _ := json.Marshal(customer)
	req, _ := http.NewRequest("POST", "/customers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response models.Customer
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %s", response.Name)
	}

	if response.ID == 0 {
		t.Error("Expected non-zero customer ID")
	}
}

func TestCreateOrder(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := setupRouter(pool)

	// First create a customer
	customer := map[string]string{
		"name":  "Jane Doe",
		"email": "jane@example.com",
	}

	body, _ := json.Marshal(customer)
	req, _ := http.NewRequest("POST", "/customers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var customerResponse models.Customer
	json.Unmarshal(w.Body.Bytes(), &customerResponse)

	// Now create an order
	order := models.CreateOrderRequest{
		CustomerID: customerResponse.ID,
		Items: []models.CreateOrderItem{
			{ProductID: 1, Quantity: 2, Price: 29.99},
			{ProductID: 2, Quantity: 1, Price: 49.99},
		},
	}

	body, _ = json.Marshal(order)
	req, _ = http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var orderResponse models.OrderWithItems
	json.Unmarshal(w.Body.Bytes(), &orderResponse)

	expectedTotal := (29.99 * 2) + (49.99 * 1)
	if orderResponse.Total != expectedTotal {
		t.Errorf("Expected total %.2f, got %.2f", expectedTotal, orderResponse.Total)
	}

	if len(orderResponse.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(orderResponse.Items))
	}
}

func TestGetOrder(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := setupRouter(pool)

	// Create customer and order
	customerID := createTestCustomer(t, router, "Test User", "test@example.com")
	orderID := createTestOrder(t, router, customerID)

	// Get the order
	req, _ := http.NewRequest("GET", fmt.Sprintf("/orders/%d", orderID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.OrderWithItems
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.ID != orderID {
		t.Errorf("Expected order ID %d, got %d", orderID, response.ID)
	}
}

func TestUpdateOrderStatus(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := setupRouter(pool)

	// Create customer and order
	customerID := createTestCustomer(t, router, "Test User", "test@example.com")
	orderID := createTestOrder(t, router, customerID)

	// Update status
	statusUpdate := models.UpdateOrderStatusRequest{
		Status: models.OrderStatusShipped,
	}

	body, _ := json.Marshal(statusUpdate)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/orders/%d/status", orderID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify the status was updated
	req, _ = http.NewRequest("GET", fmt.Sprintf("/orders/%d", orderID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response models.OrderWithItems
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Status != models.OrderStatusShipped {
		t.Errorf("Expected status 'shipped', got %s", response.Status)
	}
}

func TestOrderStats(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := setupRouter(pool)

	// Create customer and multiple orders
	customerID := createTestCustomer(t, router, "Stats User", "stats@example.com")
	createTestOrder(t, router, customerID)
	createTestOrder(t, router, customerID)

	// Get stats
	req, _ := http.NewRequest("GET", fmt.Sprintf("/customers/%d/orders/stats", customerID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &stats)

	totalOrders := int64(stats["total_orders"].(float64))
	if totalOrders != 2 {
		t.Errorf("Expected 2 total orders, got %d", totalOrders)
	}
}

// Helper functions
func createTestCustomer(t *testing.T, router *gin.Engine, name, email string) int64 {
	customer := map[string]string{
		"name":  name,
		"email": email,
	}

	body, _ := json.Marshal(customer)
	req, _ := http.NewRequest("POST", "/customers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response models.Customer
	json.Unmarshal(w.Body.Bytes(), &response)
	return response.ID
}

func createTestOrder(t *testing.T, router *gin.Engine, customerID int64) int64 {
	order := models.CreateOrderRequest{
		CustomerID: customerID,
		Items: []models.CreateOrderItem{
			{ProductID: 1, Quantity: 1, Price: 99.99},
		},
	}

	body, _ := json.Marshal(order)
	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response models.OrderWithItems
	json.Unmarshal(w.Body.Bytes(), &response)
	return response.ID
}

func getEnvOrDefault(key, defaultValue string) string {
	// In a real implementation, this would check os.Getenv
	return defaultValue
}

// Benchmark tests to generate load for OBI testing
func BenchmarkCreateOrder(b *testing.B) {
	pool := setupTestDB(&testing.T{})
	defer pool.Close()

	router := setupRouter(pool)
	customerID := createTestCustomer(&testing.T{}, router, "Bench User", fmt.Sprintf("bench%d@example.com", time.Now().Unix()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order := models.CreateOrderRequest{
			CustomerID: customerID,
			Items: []models.CreateOrderItem{
				{ProductID: int64(i % 100), Quantity: i%10 + 1, Price: float64(i%100) + 9.99},
			},
		}

		body, _ := json.Marshal(order)
		req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
