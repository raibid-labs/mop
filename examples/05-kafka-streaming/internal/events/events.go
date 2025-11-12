package events

import (
	"encoding/json"
	"time"
)

// OrderEventType represents the type of order event
type OrderEventType string

const (
	OrderCreated   OrderEventType = "order.created"
	OrderUpdated   OrderEventType = "order.updated"
	OrderCancelled OrderEventType = "order.cancelled"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

// OrderEvent represents an event related to an order
type OrderEvent struct {
	ID          string         `json:"id"`
	Type        OrderEventType `json:"type"`
	OrderID     string         `json:"order_id"`
	CustomerID  string         `json:"customer_id"`
	Status      OrderStatus    `json:"status"`
	Total       float64        `json:"total"`
	Items       []OrderItem    `json:"items,omitempty"`
	Timestamp   time.Time      `json:"timestamp"`
	Version     int            `json:"version"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// NotificationEvent represents a notification to be sent
type NotificationEvent struct {
	ID          string         `json:"id"`
	OrderID     string         `json:"order_id"`
	CustomerID  string         `json:"customer_id"`
	Type        string         `json:"type"`
	Message     string         `json:"message"`
	Email       string         `json:"email"`
	Timestamp   time.Time      `json:"timestamp"`
}

// AnalyticsEvent represents an analytics event
type AnalyticsEvent struct {
	ID          string         `json:"id"`
	EventType   string         `json:"event_type"`
	OrderID     string         `json:"order_id"`
	CustomerID  string         `json:"customer_id"`
	Total       float64        `json:"total"`
	ItemCount   int            `json:"item_count"`
	Status      OrderStatus    `json:"status"`
	Timestamp   time.Time      `json:"timestamp"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// Marshal serializes an event to JSON bytes
func Marshal(event any) ([]byte, error) {
	return json.Marshal(event)
}

// UnmarshalOrderEvent deserializes JSON bytes to an OrderEvent
func UnmarshalOrderEvent(data []byte) (*OrderEvent, error) {
	var event OrderEvent
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// UnmarshalNotificationEvent deserializes JSON bytes to a NotificationEvent
func UnmarshalNotificationEvent(data []byte) (*NotificationEvent, error) {
	var event NotificationEvent
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// UnmarshalAnalyticsEvent deserializes JSON bytes to an AnalyticsEvent
func UnmarshalAnalyticsEvent(data []byte) (*AnalyticsEvent, error) {
	var event AnalyticsEvent
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GetPartitionKey returns the key for partitioning messages
func (e *OrderEvent) GetPartitionKey() string {
	return e.OrderID
}

// Validate checks if the order event is valid
func (e *OrderEvent) Validate() error {
	if e.ID == "" {
		return ErrMissingEventID
	}
	if e.OrderID == "" {
		return ErrMissingOrderID
	}
	if e.CustomerID == "" {
		return ErrMissingCustomerID
	}
	if e.Type == "" {
		return ErrMissingEventType
	}
	if e.Version < 1 {
		return ErrInvalidVersion
	}
	return nil
}
