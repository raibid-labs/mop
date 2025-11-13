package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raibid-labs/mop/examples/03-sql-app/internal/models"
)

// OrderRepository handles order data operations
type OrderRepository struct {
	db *pgxpool.Pool
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create creates a new order with items in a transaction
func (r *OrderRepository) Create(ctx context.Context, req *models.CreateOrderRequest) (*models.OrderWithItems, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Calculate total
	var total float64
	for _, item := range req.Items {
		total += item.Price * float64(item.Quantity)
	}

	// Create order
	orderQuery := `
		INSERT INTO orders (customer_id, status, total)
		VALUES ($1, $2, $3)
		RETURNING id, customer_id, status, total, created_at, updated_at
	`

	var order models.Order
	err = tx.QueryRow(ctx, orderQuery, req.CustomerID, models.OrderStatusPending, total).
		Scan(&order.ID, &order.CustomerID, &order.Status, &order.Total, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items
	items := make([]models.OrderItem, 0, len(req.Items))
	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, order_id, product_id, quantity, price, created_at
	`

	for _, reqItem := range req.Items {
		var item models.OrderItem
		err = tx.QueryRow(ctx, itemQuery, order.ID, reqItem.ProductID, reqItem.Quantity, reqItem.Price).
			Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}
		items = append(items, item)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &models.OrderWithItems{
		Order: order,
		Items: items,
	}, nil
}

// GetByID retrieves an order by ID with its items
func (r *OrderRepository) GetByID(ctx context.Context, id int64) (*models.OrderWithItems, error) {
	// Get order
	orderQuery := `
		SELECT id, customer_id, status, total, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order models.Order
	err := r.db.QueryRow(ctx, orderQuery, id).
		Scan(&order.ID, &order.CustomerID, &order.Status, &order.Total, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get order items
	itemsQuery := `
		SELECT id, order_id, product_id, quantity, price, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	return &models.OrderWithItems{
		Order: order,
		Items: items,
	}, nil
}

// GetByIDWithCustomer retrieves an order with customer details (demonstrates JOIN)
func (r *OrderRepository) GetByIDWithCustomer(ctx context.Context, id int64) (*models.OrderWithItems, error) {
	// This query demonstrates a JOIN that will be captured by OBI
	orderQuery := `
		SELECT
			o.id, o.customer_id, o.status, o.total, o.created_at, o.updated_at,
			c.id, c.name, c.email, c.created_at
		FROM orders o
		INNER JOIN customers c ON o.customer_id = c.id
		WHERE o.id = $1
	`

	var order models.Order
	var customer models.Customer
	err := r.db.QueryRow(ctx, orderQuery, id).
		Scan(
			&order.ID, &order.CustomerID, &order.Status, &order.Total, &order.CreatedAt, &order.UpdatedAt,
			&customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt,
		)
	if err != nil {
		return nil, fmt.Errorf("failed to get order with customer: %w", err)
	}

	// Get order items
	itemsQuery := `
		SELECT id, order_id, product_id, quantity, price, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	return &models.OrderWithItems{
		Order:    order,
		Items:    items,
		Customer: &customer,
	}, nil
}

// ListByCustomer retrieves all orders for a customer
func (r *OrderRepository) ListByCustomer(ctx context.Context, customerID int64, limit, offset int) ([]models.Order, error) {
	query := `
		SELECT id, customer_id, status, total, created_at, updated_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, customerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.CustomerID, &order.Status, &order.Total, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

// UpdateStatus updates the status of an order
func (r *OrderRepository) UpdateStatus(ctx context.Context, id int64, status models.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetOrderStats retrieves order statistics (demonstrates aggregation)
func (r *OrderRepository) GetOrderStats(ctx context.Context, customerID int64) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_orders,
			COALESCE(SUM(total), 0) as total_spent,
			COALESCE(AVG(total), 0) as average_order_value,
			COUNT(CASE WHEN status = 'delivered' THEN 1 END) as delivered_orders,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_orders
		FROM orders
		WHERE customer_id = $1
	`

	var stats struct {
		TotalOrders       int64
		TotalSpent        float64
		AverageOrderValue float64
		DeliveredOrders   int64
		CancelledOrders   int64
	}

	err := r.db.QueryRow(ctx, query, customerID).
		Scan(&stats.TotalOrders, &stats.TotalSpent, &stats.AverageOrderValue, &stats.DeliveredOrders, &stats.CancelledOrders)
	if err != nil {
		return nil, fmt.Errorf("failed to get order stats: %w", err)
	}

	return map[string]interface{}{
		"total_orders":         stats.TotalOrders,
		"total_spent":          stats.TotalSpent,
		"average_order_value":  stats.AverageOrderValue,
		"delivered_orders":     stats.DeliveredOrders,
		"cancelled_orders":     stats.CancelledOrders,
	}, nil
}

// SimulateSlowQuery simulates a slow query for OBI testing (N+1 problem)
func (r *OrderRepository) SimulateSlowQuery(ctx context.Context, customerID int64) ([]models.OrderWithItems, error) {
	// This is intentionally inefficient - demonstrates N+1 query problem
	// OBI should capture all these queries

	orders, err := r.ListByCustomer(ctx, customerID, 100, 0)
	if err != nil {
		return nil, err
	}

	result := make([]models.OrderWithItems, 0, len(orders))
	for _, order := range orders {
		// N+1: One query per order to get items
		orderWithItems, err := r.GetByID(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, *orderWithItems)
	}

	return result, nil
}
