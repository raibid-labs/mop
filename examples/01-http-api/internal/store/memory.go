package store

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raibid-labs/mop/examples/01-http-api/internal/models"
)

var (
	// ErrNotFound is returned when a product is not found
	ErrNotFound = errors.New("product not found")
)

// Store defines the interface for product storage
type Store interface {
	Create(product *models.Product) error
	Get(id string) (*models.Product, error)
	Update(id string, product *models.Product) error
	Delete(id string) error
	List(limit, offset int) ([]models.Product, int, error)
	Search(query string, limit, offset int) ([]models.Product, int, error)
}

// MemoryStore implements Store using an in-memory map
type MemoryStore struct {
	mu       sync.RWMutex
	products map[string]*models.Product
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		products: make(map[string]*models.Product),
	}
}

// Create adds a new product to the store
func (s *MemoryStore) Create(product *models.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	product.ID = uuid.New().String()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	s.products[product.ID] = product
	return nil
}

// Get retrieves a product by ID
func (s *MemoryStore) Get(id string) (*models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, exists := s.products[id]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy to prevent external modifications
	productCopy := *product
	return &productCopy, nil
}

// Update modifies an existing product
func (s *MemoryStore) Update(id string, product *models.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.products[id]
	if !exists {
		return ErrNotFound
	}

	// Keep original ID and CreatedAt
	product.ID = existing.ID
	product.CreatedAt = existing.CreatedAt
	product.UpdatedAt = time.Now()

	s.products[id] = product
	return nil
}

// Delete removes a product from the store
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		return ErrNotFound
	}

	delete(s.products, id)
	return nil
}

// List returns a paginated list of all products
func (s *MemoryStore) List(limit, offset int) ([]models.Product, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.products)
	products := make([]models.Product, 0, limit)

	// Convert map to slice
	allProducts := make([]models.Product, 0, total)
	for _, p := range s.products {
		allProducts = append(allProducts, *p)
	}

	// Apply pagination
	start := offset
	if start > len(allProducts) {
		start = len(allProducts)
	}

	end := start + limit
	if end > len(allProducts) {
		end = len(allProducts)
	}

	products = allProducts[start:end]

	return products, total, nil
}

// Search finds products by name or description (case-insensitive substring match)
func (s *MemoryStore) Search(query string, limit, offset int) ([]models.Product, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	matchingProducts := make([]models.Product, 0)

	for _, p := range s.products {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Description), query) {
			matchingProducts = append(matchingProducts, *p)
		}
	}

	total := len(matchingProducts)

	// Apply pagination
	start := offset
	if start > len(matchingProducts) {
		start = len(matchingProducts)
	}

	end := start + limit
	if end > len(matchingProducts) {
		end = len(matchingProducts)
	}

	return matchingProducts[start:end], total, nil
}
