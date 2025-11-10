package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/raibid-labs/mop/examples/01-http-api/internal/models"
	"github.com/raibid-labs/mop/examples/01-http-api/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest() (*gin.Engine, *store.MemoryStore) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	st := store.NewMemoryStore()
	return r, st
}

func TestProductHandler_Create(t *testing.T) {
	r, st := setupTest()
	handler := NewProductHandler(st)

	r.POST("/products", handler.Create)

	t.Run("valid product", func(t *testing.T) {
		product := map[string]interface{}{
			"name":        "Test Product",
			"description": "Test Description",
			"price":       99.99,
			"stock":       100,
		}

		body, _ := json.Marshal(product)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.Product
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.ID)
		assert.Equal(t, "Test Product", response.Name)
	})

	t.Run("invalid product - missing name", func(t *testing.T) {
		product := map[string]interface{}{
			"price": 99.99,
			"stock": 100,
		}

		body, _ := json.Marshal(product)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid product - negative price", func(t *testing.T) {
		product := map[string]interface{}{
			"name":  "Test Product",
			"price": -10.0,
			"stock": 100,
		}

		body, _ := json.Marshal(product)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestProductHandler_Get(t *testing.T) {
	r, st := setupTest()
	handler := NewProductHandler(st)

	r.GET("/products/:id", handler.Get)

	// Create a product
	product := &models.Product{
		Name:  "Test Product",
		Price: 99.99,
		Stock: 100,
	}
	err := st.Create(product)
	require.NoError(t, err)

	t.Run("existing product", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products/"+product.ID, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Product
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, product.ID, response.ID)
		assert.Equal(t, "Test Product", response.Name)
	})

	t.Run("non-existent product", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products/non-existent", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestProductHandler_List(t *testing.T) {
	r, st := setupTest()
	handler := NewProductHandler(st)

	r.GET("/products", handler.List)

	// Create multiple products
	for i := 0; i < 15; i++ {
		product := &models.Product{
			Name:  "Product",
			Price: 99.99,
			Stock: 100,
		}
		err := st.Create(product)
		require.NoError(t, err)
	}

	t.Run("default pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 15, response.Total)
		assert.Equal(t, 10, response.Limit)
		assert.Len(t, response.Products, 10)
	})

	t.Run("custom pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products?limit=5&offset=10", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 15, response.Total)
		assert.Equal(t, 5, response.Limit)
		assert.Len(t, response.Products, 5)
	})
}

func TestProductHandler_Update(t *testing.T) {
	r, st := setupTest()
	handler := NewProductHandler(st)

	r.PUT("/products/:id", handler.Update)

	// Create a product
	product := &models.Product{
		Name:  "Original Name",
		Price: 99.99,
		Stock: 100,
	}
	err := st.Create(product)
	require.NoError(t, err)

	t.Run("valid update", func(t *testing.T) {
		update := map[string]interface{}{
			"name":  "Updated Name",
			"price": 149.99,
			"stock": 50,
		}

		body, _ := json.Marshal(update)
		req := httptest.NewRequest(http.MethodPut, "/products/"+product.ID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Product
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", response.Name)
		assert.Equal(t, 149.99, response.Price)
	})

	t.Run("non-existent product", func(t *testing.T) {
		update := map[string]interface{}{
			"name":  "Updated Name",
			"price": 149.99,
			"stock": 50,
		}

		body, _ := json.Marshal(update)
		req := httptest.NewRequest(http.MethodPut, "/products/non-existent", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestProductHandler_Delete(t *testing.T) {
	r, st := setupTest()
	handler := NewProductHandler(st)

	r.DELETE("/products/:id", handler.Delete)

	// Create a product
	product := &models.Product{
		Name:  "Test Product",
		Price: 99.99,
		Stock: 100,
	}
	err := st.Create(product)
	require.NoError(t, err)

	t.Run("existing product", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/products/"+product.ID, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify deletion
		_, err := st.Get(product.ID)
		assert.Equal(t, store.ErrNotFound, err)
	})

	t.Run("non-existent product", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/products/non-existent", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestProductHandler_Search(t *testing.T) {
	r, st := setupTest()
	handler := NewProductHandler(st)

	r.GET("/search", handler.Search)

	// Create products
	products := []*models.Product{
		{Name: "Apple iPhone", Description: "Smartphone", Price: 999.99, Stock: 10},
		{Name: "Samsung Galaxy", Description: "Smartphone", Price: 899.99, Stock: 15},
		{Name: "Apple MacBook", Description: "Laptop", Price: 1999.99, Stock: 5},
	}

	for _, p := range products {
		err := st.Create(p)
		require.NoError(t, err)
	}

	t.Run("search with results", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/search?q=Apple", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 2, response.Total)
		assert.Len(t, response.Products, 2)
	})

	t.Run("search with no results", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/search?q=NonExistent", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 0, response.Total)
		assert.Len(t, response.Products, 0)
	})

	t.Run("missing query parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/search", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
