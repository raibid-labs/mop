package models

import "time"

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// Order represents an order in the system
type Order struct {
	ID         int64       `json:"id" db:"id"`
	CustomerID int64       `json:"customer_id" db:"customer_id"`
	Status     OrderStatus `json:"status" db:"status"`
	Total      float64     `json:"total" db:"total"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        int64     `json:"id" db:"id"`
	OrderID   int64     `json:"order_id" db:"order_id"`
	ProductID int64     `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Price     float64   `json:"price" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// OrderWithItems represents an order with its items
type OrderWithItems struct {
	Order
	Items    []OrderItem `json:"items"`
	Customer *Customer   `json:"customer,omitempty"`
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	CustomerID int64              `json:"customer_id" binding:"required"`
	Items      []CreateOrderItem  `json:"items" binding:"required,min=1"`
}

// CreateOrderItem represents an item in the create order request
type CreateOrderItem struct {
	ProductID int64   `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	Price     float64 `json:"price" binding:"required,min=0"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}
