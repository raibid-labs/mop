package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raibid-labs/mop/examples/03-sql-app/internal/models"
)

// CustomerRepository handles customer data operations
type CustomerRepository struct {
	db *pgxpool.Pool
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(db *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// Create creates a new customer
func (r *CustomerRepository) Create(ctx context.Context, customer *models.Customer) error {
	query := `
		INSERT INTO customers (name, email)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query, customer.Name, customer.Email).
		Scan(&customer.ID, &customer.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

// GetByID retrieves a customer by ID
func (r *CustomerRepository) GetByID(ctx context.Context, id int64) (*models.Customer, error) {
	query := `
		SELECT id, name, email, created_at
		FROM customers
		WHERE id = $1
	`

	var customer models.Customer
	err := r.db.QueryRow(ctx, query, id).
		Scan(&customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

// GetByEmail retrieves a customer by email
func (r *CustomerRepository) GetByEmail(ctx context.Context, email string) (*models.Customer, error) {
	query := `
		SELECT id, name, email, created_at
		FROM customers
		WHERE email = $1
	`

	var customer models.Customer
	err := r.db.QueryRow(ctx, query, email).
		Scan(&customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return &customer, nil
}

// List retrieves all customers with pagination
func (r *CustomerRepository) List(ctx context.Context, limit, offset int) ([]models.Customer, error) {
	query := `
		SELECT id, name, email, created_at
		FROM customers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		if err := rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %w", err)
	}

	return customers, nil
}
