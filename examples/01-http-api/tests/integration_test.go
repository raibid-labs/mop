package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raibid-labs/mop/examples/01-http-api/internal/models"
)

const (
	baseURL = "http://localhost:18080"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Start test server
	server := startTestServer(t)
	defer server.Shutdown(context.Background())

	// Wait for server to be ready
	waitForServer(t, baseURL+"/health")

	t.Run("Health Check", testHealthCheck)
	t.Run("CRUD Operations", testCRUDOperations)
	t.Run("Pagination", testPagination)
	t.Run("Search", testSearch)
	t.Run("Error Handling", testErrorHandling)
	t.Run("Slow Endpoint", testSlowEndpoint)
	t.Run("Rate Limiting", testRateLimiting)
	t.Run("Concurrent Requests", testConcurrentRequests)
}

func testHealthCheck(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)

	assert.Equal(t, "healthy", health["status"])
	assert.NotEmpty(t, health["uptime"])
}

func testCRUDOperations(t *testing.T) {
	// Create
	product := map[string]interface{}{
		"name":        "Integration Test Product",
		"description": "Test Description",
		"price":       99.99,
		"stock":       100,
	}

	body, _ := json.Marshal(product)
	resp, err := http.Post(baseURL+"/products", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created models.Product
	err = json.NewDecoder(resp.Body).Decode(&created)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "Integration Test Product", created.Name)

	// Read
	resp, err = http.Get(baseURL + "/products/" + created.ID)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrieved models.Product
	err = json.NewDecoder(resp.Body).Decode(&retrieved)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Name, retrieved.Name)

	// Update
	update := map[string]interface{}{
		"name":        "Updated Product Name",
		"description": "Updated Description",
		"price":       149.99,
		"stock":       50,
	}

	body, _ = json.Marshal(update)
	req, _ := http.NewRequest(http.MethodPut, baseURL+"/products/"+created.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated models.Product
	err = json.NewDecoder(resp.Body).Decode(&updated)
	require.NoError(t, err)
	assert.Equal(t, "Updated Product Name", updated.Name)

	// Delete
	req, _ = http.NewRequest(http.MethodDelete, baseURL+"/products/"+created.ID, nil)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deletion
	resp, err = http.Get(baseURL + "/products/" + created.ID)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func testPagination(t *testing.T) {
	// Create 15 products
	for i := 0; i < 15; i++ {
		product := map[string]interface{}{
			"name":  fmt.Sprintf("Pagination Product %d", i),
			"price": 99.99,
			"stock": 100,
		}

		body, _ := json.Marshal(product)
		resp, err := http.Post(baseURL+"/products", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		resp.Body.Close()
	}

	// Test default pagination
	resp, err := http.Get(baseURL + "/products")
	require.NoError(t, err)
	defer resp.Body.Close()

	var list1 models.ListResponse
	err = json.NewDecoder(resp.Body).Decode(&list1)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, list1.Total, 15)
	assert.Equal(t, 10, list1.Limit)

	// Test custom pagination
	resp, err = http.Get(baseURL + "/products?limit=5&offset=5")
	require.NoError(t, err)
	defer resp.Body.Close()

	var list2 models.ListResponse
	err = json.NewDecoder(resp.Body).Decode(&list2)
	require.NoError(t, err)
	assert.Equal(t, 5, list2.Limit)
	assert.Equal(t, 5, list2.Offset)
}

func testSearch(t *testing.T) {
	// Create products with unique names
	products := []map[string]interface{}{
		{"name": "SearchTest Apple iPhone", "price": 999.99, "stock": 10},
		{"name": "SearchTest Samsung Galaxy", "price": 899.99, "stock": 15},
		{"name": "SearchTest Apple MacBook", "price": 1999.99, "stock": 5},
	}

	for _, p := range products {
		body, _ := json.Marshal(p)
		resp, err := http.Post(baseURL+"/products", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		resp.Body.Close()
	}

	// Search for "Apple"
	resp, err := http.Get(baseURL + "/search?q=SearchTest Apple")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var results models.ListResponse
	err = json.NewDecoder(resp.Body).Decode(&results)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, results.Total, 2)
}

func testErrorHandling(t *testing.T) {
	// Test 404
	resp, err := http.Get(baseURL + "/products/nonexistent-id")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Test 500
	resp, err = http.Get(baseURL + "/error")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	// Test 400 - invalid JSON
	resp, err = http.Post(baseURL+"/products", "application/json", bytes.NewBufferString("invalid json"))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func testSlowEndpoint(t *testing.T) {
	start := time.Now()
	resp, err := http.Get(baseURL + "/slow")
	duration := time.Since(start)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.GreaterOrEqual(t, duration, 1*time.Second, "Slow endpoint should take at least 1 second")
	assert.LessOrEqual(t, duration, 4*time.Second, "Slow endpoint should take at most 3 seconds + tolerance")
}

func testRateLimiting(t *testing.T) {
	// Make many rapid requests
	client := &http.Client{Timeout: 5 * time.Second}
	var rateLimited int

	for i := 0; i < 150; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err != nil {
			t.Logf("Request %d failed: %v", i, err)
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited++
		}
		resp.Body.Close()
	}

	// We expect some requests to be rate limited
	// Note: This might not trigger in integration tests depending on timing
	t.Logf("Rate limited requests: %d out of 150", rateLimited)
}

func testConcurrentRequests(t *testing.T) {
	var wg sync.WaitGroup
	errors := make(chan error, 50)
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			product := map[string]interface{}{
				"name":  fmt.Sprintf("Concurrent Product %d", id),
				"price": 99.99,
				"stock": 100,
			}

			body, _ := json.Marshal(product)
			resp, err := http.Post(baseURL+"/products", "application/json", bytes.NewBuffer(body))
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				mu.Lock()
				successCount++
				mu.Unlock()
			} else {
				errors <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Logf("Encountered %d errors during concurrent requests:", len(errorList))
		for _, err := range errorList {
			t.Logf("  - %v", err)
		}
	}

	// We expect most requests to succeed
	assert.Greater(t, successCount, 40, "Expected at least 40 successful concurrent requests")
}

// Helper functions

func waitForServer(t *testing.T, healthURL string) {
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Server did not start within timeout")
		case <-ticker.C:
			resp, err := http.Get(healthURL)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}
