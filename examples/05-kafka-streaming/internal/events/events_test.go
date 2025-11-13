package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   OrderEvent
		wantErr error
	}{
		{
			name: "valid event",
			event: OrderEvent{
				ID:         "evt-123",
				Type:       OrderCreated,
				OrderID:    "ord-123",
				CustomerID: "cust-123",
				Status:     StatusPending,
				Total:      99.99,
				Timestamp:  time.Now(),
				Version:    1,
			},
			wantErr: nil,
		},
		{
			name: "missing event ID",
			event: OrderEvent{
				Type:       OrderCreated,
				OrderID:    "ord-123",
				CustomerID: "cust-123",
				Version:    1,
			},
			wantErr: ErrMissingEventID,
		},
		{
			name: "missing order ID",
			event: OrderEvent{
				ID:         "evt-123",
				Type:       OrderCreated,
				CustomerID: "cust-123",
				Version:    1,
			},
			wantErr: ErrMissingOrderID,
		},
		{
			name: "missing customer ID",
			event: OrderEvent{
				ID:      "evt-123",
				Type:    OrderCreated,
				OrderID: "ord-123",
				Version: 1,
			},
			wantErr: ErrMissingCustomerID,
		},
		{
			name: "missing event type",
			event: OrderEvent{
				ID:         "evt-123",
				OrderID:    "ord-123",
				CustomerID: "cust-123",
				Version:    1,
			},
			wantErr: ErrMissingEventType,
		},
		{
			name: "invalid version",
			event: OrderEvent{
				ID:         "evt-123",
				Type:       OrderCreated,
				OrderID:    "ord-123",
				CustomerID: "cust-123",
				Version:    0,
			},
			wantErr: ErrInvalidVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderEvent_GetPartitionKey(t *testing.T) {
	event := OrderEvent{
		OrderID: "ord-123",
	}
	assert.Equal(t, "ord-123", event.GetPartitionKey())
}

func TestMarshalUnmarshalOrderEvent(t *testing.T) {
	original := &OrderEvent{
		ID:         "evt-123",
		Type:       OrderCreated,
		OrderID:    "ord-123",
		CustomerID: "cust-123",
		Status:     StatusPending,
		Total:      99.99,
		Items: []OrderItem{
			{
				ProductID: "prod-1",
				Name:      "Product 1",
				Quantity:  2,
				Price:     49.99,
			},
		},
		Timestamp: time.Date(2023, 11, 10, 12, 0, 0, 0, time.UTC),
		Version:   1,
		Metadata: map[string]any{
			"source": "api",
			"user":   "admin",
		},
	}

	// Marshal
	data, err := Marshal(original)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal
	decoded, err := UnmarshalOrderEvent(data)
	require.NoError(t, err)
	require.NotNil(t, decoded)

	// Verify
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Type, decoded.Type)
	assert.Equal(t, original.OrderID, decoded.OrderID)
	assert.Equal(t, original.CustomerID, decoded.CustomerID)
	assert.Equal(t, original.Status, decoded.Status)
	assert.Equal(t, original.Total, decoded.Total)
	assert.Equal(t, original.Version, decoded.Version)
	assert.Equal(t, len(original.Items), len(decoded.Items))
	assert.Equal(t, original.Items[0].ProductID, decoded.Items[0].ProductID)
}

func TestMarshalUnmarshalNotificationEvent(t *testing.T) {
	original := &NotificationEvent{
		ID:         "notif-123",
		OrderID:    "ord-123",
		CustomerID: "cust-123",
		Type:       "email",
		Message:    "Your order has been created",
		Email:      "customer@example.com",
		Timestamp:  time.Date(2023, 11, 10, 12, 0, 0, 0, time.UTC),
	}

	// Marshal
	data, err := Marshal(original)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal
	decoded, err := UnmarshalNotificationEvent(data)
	require.NoError(t, err)
	require.NotNil(t, decoded)

	// Verify
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.OrderID, decoded.OrderID)
	assert.Equal(t, original.CustomerID, decoded.CustomerID)
	assert.Equal(t, original.Type, decoded.Type)
	assert.Equal(t, original.Message, decoded.Message)
	assert.Equal(t, original.Email, decoded.Email)
}

func TestMarshalUnmarshalAnalyticsEvent(t *testing.T) {
	original := &AnalyticsEvent{
		ID:         "analytics-123",
		EventType:  "order_created",
		OrderID:    "ord-123",
		CustomerID: "cust-123",
		Total:      99.99,
		ItemCount:  2,
		Status:     StatusPending,
		Timestamp:  time.Date(2023, 11, 10, 12, 0, 0, 0, time.UTC),
		Metadata: map[string]any{
			"region": "US",
		},
	}

	// Marshal
	data, err := Marshal(original)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal
	decoded, err := UnmarshalAnalyticsEvent(data)
	require.NoError(t, err)
	require.NotNil(t, decoded)

	// Verify
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.EventType, decoded.EventType)
	assert.Equal(t, original.OrderID, decoded.OrderID)
	assert.Equal(t, original.CustomerID, decoded.CustomerID)
	assert.Equal(t, original.Total, decoded.Total)
	assert.Equal(t, original.ItemCount, decoded.ItemCount)
	assert.Equal(t, original.Status, decoded.Status)
}

func TestUnmarshalOrderEvent_InvalidJSON(t *testing.T) {
	_, err := UnmarshalOrderEvent([]byte("invalid json"))
	assert.Error(t, err)
}

func TestUnmarshalNotificationEvent_InvalidJSON(t *testing.T) {
	_, err := UnmarshalNotificationEvent([]byte("invalid json"))
	assert.Error(t, err)
}

func TestUnmarshalAnalyticsEvent_InvalidJSON(t *testing.T) {
	_, err := UnmarshalAnalyticsEvent([]byte("invalid json"))
	assert.Error(t, err)
}
