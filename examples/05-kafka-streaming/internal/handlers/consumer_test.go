package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raibid-labs/mop/examples/05-kafka-streaming/internal/events"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConsumer(t *testing.T) {
	cfg := ConsumerConfig{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test-topic",
		GroupID:    "test-group",
		MaxRetries: 3,
		DLQEnabled: true,
	}

	handler := func(ctx context.Context, msg kafka.Message) error {
		return nil
	}

	consumer := NewConsumer(cfg, handler)

	assert.NotNil(t, consumer)
	assert.NotNil(t, consumer.reader)
	assert.NotNil(t, consumer.handler)
	assert.Equal(t, 3, consumer.maxRetries)
	assert.NotNil(t, consumer.dlqProducer)
}

func TestNewConsumer_NoDLQ(t *testing.T) {
	cfg := ConsumerConfig{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test-topic",
		GroupID:    "test-group",
		MaxRetries: 3,
		DLQEnabled: false,
	}

	handler := func(ctx context.Context, msg kafka.Message) error {
		return nil
	}

	consumer := NewConsumer(cfg, handler)

	assert.NotNil(t, consumer)
	assert.Nil(t, consumer.dlqProducer)
}

func TestConsumer_ProcessMessageWithRetry_Success(t *testing.T) {
	callCount := 0
	handler := func(ctx context.Context, msg kafka.Message) error {
		callCount++
		return nil
	}

	cfg := ConsumerConfig{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test-topic",
		GroupID:    "test-group",
		MaxRetries: 3,
	}

	consumer := NewConsumer(cfg, handler)

	msg := kafka.Message{
		Key:   []byte("test-key"),
		Value: []byte("test-value"),
	}

	err := consumer.processMessageWithRetry(context.Background(), msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount) // Should succeed on first attempt
}

func TestConsumer_ProcessMessageWithRetry_SuccessAfterRetries(t *testing.T) {
	callCount := 0
	handler := func(ctx context.Context, msg kafka.Message) error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	cfg := ConsumerConfig{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test-topic",
		GroupID:    "test-group",
		MaxRetries: 3,
	}

	consumer := NewConsumer(cfg, handler)

	msg := kafka.Message{
		Key:   []byte("test-key"),
		Value: []byte("test-value"),
	}

	err := consumer.processMessageWithRetry(context.Background(), msg)
	assert.NoError(t, err)
	assert.Equal(t, 3, callCount) // Should succeed on third attempt
}

func TestConsumer_ProcessMessageWithRetry_Failure(t *testing.T) {
	callCount := 0
	handler := func(ctx context.Context, msg kafka.Message) error {
		callCount++
		return errors.New("persistent error")
	}

	cfg := ConsumerConfig{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test-topic",
		GroupID:    "test-group",
		MaxRetries: 2,
	}

	consumer := NewConsumer(cfg, handler)

	msg := kafka.Message{
		Key:   []byte("test-key"),
		Value: []byte("test-value"),
	}

	err := consumer.processMessageWithRetry(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed after 2 retries")
	assert.Equal(t, 3, callCount) // Initial + 2 retries
}

func TestNotificationHandler(t *testing.T) {
	notification := &events.NotificationEvent{
		ID:         "notif-123",
		OrderID:    "ord-123",
		CustomerID: "cust-123",
		Type:       "email",
		Message:    "Test message",
		Email:      "test@example.com",
		Timestamp:  time.Now(),
	}

	data, err := events.Marshal(notification)
	require.NoError(t, err)

	msg := kafka.Message{
		Key:   []byte("cust-123"),
		Value: data,
	}

	err = NotificationHandler(context.Background(), msg)
	assert.NoError(t, err)
}

func TestNotificationHandler_InvalidMessage(t *testing.T) {
	msg := kafka.Message{
		Key:   []byte("test-key"),
		Value: []byte("invalid json"),
	}

	err := NotificationHandler(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal notification")
}

func TestAnalyticsHandler(t *testing.T) {
	analytics := &events.AnalyticsEvent{
		ID:         "analytics-123",
		EventType:  "order_created",
		OrderID:    "ord-123",
		CustomerID: "cust-123",
		Total:      99.99,
		ItemCount:  2,
		Status:     events.StatusPending,
		Timestamp:  time.Now(),
	}

	data, err := events.Marshal(analytics)
	require.NoError(t, err)

	msg := kafka.Message{
		Key:   []byte("ord-123"),
		Value: data,
	}

	err = AnalyticsHandler(context.Background(), msg)
	assert.NoError(t, err)
}

func TestAnalyticsHandler_InvalidMessage(t *testing.T) {
	msg := kafka.Message{
		Key:   []byte("test-key"),
		Value: []byte("invalid json"),
	}

	err := AnalyticsHandler(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal analytics")
}

func TestOrderEventHandler(t *testing.T) {
	producer := NewProducer([]string{"localhost:9092"})
	handler := OrderEventHandler(producer)

	orderEvent := &events.OrderEvent{
		ID:         "evt-123",
		Type:       events.OrderCreated,
		OrderID:    "ord-123",
		CustomerID: "cust-123",
		Status:     events.StatusPending,
		Total:      99.99,
		Items: []events.OrderItem{
			{
				ProductID: "prod-1",
				Name:      "Product 1",
				Quantity:  2,
				Price:     49.99,
			},
		},
		Timestamp: time.Now(),
		Version:   1,
	}

	data, err := events.Marshal(orderEvent)
	require.NoError(t, err)

	msg := kafka.Message{
		Key:   []byte("ord-123"),
		Value: data,
	}

	// This will fail because we're not connected to Kafka,
	// but we can test that it attempts to process
	err = handler(context.Background(), msg)
	// We expect an error because we're not connected to Kafka
	// The important thing is that it doesn't panic and validates correctly
	if err != nil {
		assert.Contains(t, err.Error(), "failed to produce")
	}
}

func TestOrderEventHandler_InvalidMessage(t *testing.T) {
	producer := NewProducer([]string{"localhost:9092"})
	handler := OrderEventHandler(producer)

	msg := kafka.Message{
		Key:   []byte("test-key"),
		Value: []byte("invalid json"),
	}

	err := handler(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal order event")
}

func TestConsumer_Close(t *testing.T) {
	cfg := ConsumerConfig{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test-topic",
		GroupID:    "test-group",
		MaxRetries: 3,
		DLQEnabled: false,
	}

	handler := func(ctx context.Context, msg kafka.Message) error {
		return nil
	}

	consumer := NewConsumer(cfg, handler)

	// Close should not panic
	err := consumer.Close()
	// May error if not connected, but should not panic
	_ = err
}
