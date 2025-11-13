package store

import (
	"testing"

	"github.com/raibid-labs/mop/examples/01-http-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_Create(t *testing.T) {
	store := NewMemoryStore()

	product := &models.Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Stock:       100,
	}

	err := store.Create(product)
	require.NoError(t, err)
	assert.NotEmpty(t, product.ID)
	assert.NotZero(t, product.CreatedAt)
	assert.NotZero(t, product.UpdatedAt)
}

func TestMemoryStore_Get(t *testing.T) {
	store := NewMemoryStore()

	// Create a product
	product := &models.Product{
		Name:  "Test Product",
		Price: 99.99,
		Stock: 100,
	}
	err := store.Create(product)
	require.NoError(t, err)

	// Get the product
	retrieved, err := store.Get(product.ID)
	require.NoError(t, err)
	assert.Equal(t, product.Name, retrieved.Name)
	assert.Equal(t, product.Price, retrieved.Price)

	// Get non-existent product
	_, err = store.Get("non-existent-id")
	assert.Equal(t, ErrNotFound, err)
}

func TestMemoryStore_Update(t *testing.T) {
	store := NewMemoryStore()

	// Create a product
	product := &models.Product{
		Name:  "Original Name",
		Price: 99.99,
		Stock: 100,
	}
	err := store.Create(product)
	require.NoError(t, err)

	// Update the product
	updatedProduct := &models.Product{
		Name:  "Updated Name",
		Price: 149.99,
		Stock: 50,
	}
	err = store.Update(product.ID, updatedProduct)
	require.NoError(t, err)

	// Verify update
	retrieved, err := store.Get(product.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, 149.99, retrieved.Price)

	// Update non-existent product
	err = store.Update("non-existent-id", updatedProduct)
	assert.Equal(t, ErrNotFound, err)
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore()

	// Create a product
	product := &models.Product{
		Name:  "Test Product",
		Price: 99.99,
		Stock: 100,
	}
	err := store.Create(product)
	require.NoError(t, err)

	// Delete the product
	err = store.Delete(product.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = store.Get(product.ID)
	assert.Equal(t, ErrNotFound, err)

	// Delete non-existent product
	err = store.Delete("non-existent-id")
	assert.Equal(t, ErrNotFound, err)
}

func TestMemoryStore_List(t *testing.T) {
	store := NewMemoryStore()

	// Create multiple products
	for i := 0; i < 15; i++ {
		product := &models.Product{
			Name:  "Product",
			Price: 99.99,
			Stock: 100,
		}
		err := store.Create(product)
		require.NoError(t, err)
	}

	// Test pagination
	products, total, err := store.List(10, 0)
	require.NoError(t, err)
	assert.Equal(t, 15, total)
	assert.Len(t, products, 10)

	// Test offset
	products, total, err = store.List(10, 10)
	require.NoError(t, err)
	assert.Equal(t, 15, total)
	assert.Len(t, products, 5)
}

func TestMemoryStore_Search(t *testing.T) {
	store := NewMemoryStore()

	// Create products with different names
	products := []*models.Product{
		{Name: "Apple iPhone", Description: "Smartphone", Price: 999.99, Stock: 10},
		{Name: "Samsung Galaxy", Description: "Smartphone", Price: 899.99, Stock: 15},
		{Name: "Apple MacBook", Description: "Laptop", Price: 1999.99, Stock: 5},
	}

	for _, p := range products {
		err := store.Create(p)
		require.NoError(t, err)
	}

	// Search for "Apple"
	results, total, err := store.Search("Apple", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, results, 2)

	// Search for "smartphone"
	results, total, err = store.Search("smartphone", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)

	// Search for non-existent term
	results, total, err = store.Search("nonexistent", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, results, 0)
}

func TestMemoryStore_Concurrency(t *testing.T) {
	store := NewMemoryStore()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			product := &models.Product{
				Name:  "Concurrent Product",
				Price: 99.99,
				Stock: 100,
			}
			err := store.Create(product)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all products were created
	products, total, err := store.List(20, 0)
	require.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.Len(t, products, 10)
}
