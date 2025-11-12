package events

import "errors"

var (
	// ErrMissingEventID is returned when an event has no ID
	ErrMissingEventID = errors.New("event ID is required")

	// ErrMissingOrderID is returned when an order event has no order ID
	ErrMissingOrderID = errors.New("order ID is required")

	// ErrMissingCustomerID is returned when an event has no customer ID
	ErrMissingCustomerID = errors.New("customer ID is required")

	// ErrMissingEventType is returned when an event has no type
	ErrMissingEventType = errors.New("event type is required")

	// ErrInvalidVersion is returned when an event has an invalid version
	ErrInvalidVersion = errors.New("event version must be >= 1")
)
