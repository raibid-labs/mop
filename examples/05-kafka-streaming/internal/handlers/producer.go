package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/raibid-labs/mop/examples/05-kafka-streaming/internal/events"
	"github.com/segmentio/kafka-go"
)

// Producer wraps a Kafka writer for producing messages
type Producer struct {
	writers map[string]*kafka.Writer
	brokers []string
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string) *Producer {
	return &Producer{
		writers: make(map[string]*kafka.Writer),
		brokers: brokers,
	}
}

// getWriter returns a writer for the specified topic, creating it if needed
func (p *Producer) getWriter(topic string) *kafka.Writer {
	if writer, exists := p.writers[topic]; exists {
		return writer
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{}, // Hash balancer for partition by key
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		Async:        false, // Synchronous writes for reliability
	}

	p.writers[topic] = writer
	return writer
}

// ProduceOrderEvent sends an order event to the appropriate topic
func (p *Producer) ProduceOrderEvent(ctx context.Context, event *events.OrderEvent) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	// Determine topic based on event type
	var topic string
	switch event.Type {
	case events.OrderCreated:
		topic = "orders.created"
	case events.OrderUpdated:
		topic = "orders.updated"
	case events.OrderCancelled:
		topic = "orders.cancelled"
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}

	// Serialize event
	data, err := events.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	msg := kafka.Message{
		Key:   []byte(event.GetPartitionKey()),
		Value: data,
		Time:  event.Timestamp,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.Type)},
			{Key: "event-id", Value: []byte(event.ID)},
			{Key: "version", Value: []byte(fmt.Sprintf("%d", event.Version))},
		},
	}

	// Get writer for topic
	writer := p.getWriter(topic)

	// Write message
	err = writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write message to topic %s: %w", topic, err)
	}

	log.Printf("Produced event %s to topic %s (partition key: %s)", event.ID, topic, event.GetPartitionKey())
	return nil
}

// ProduceNotification sends a notification event
func (p *Producer) ProduceNotification(ctx context.Context, notification *events.NotificationEvent) error {
	// Serialize notification
	data, err := events.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Create Kafka message
	msg := kafka.Message{
		Key:   []byte(notification.CustomerID),
		Value: data,
		Time:  notification.Timestamp,
		Headers: []kafka.Header{
			{Key: "notification-type", Value: []byte(notification.Type)},
			{Key: "notification-id", Value: []byte(notification.ID)},
		},
	}

	// Get writer for notifications topic
	writer := p.getWriter("notifications")

	// Write message
	err = writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write notification: %w", err)
	}

	log.Printf("Produced notification %s for customer %s", notification.ID, notification.CustomerID)
	return nil
}

// ProduceAnalytics sends an analytics event
func (p *Producer) ProduceAnalytics(ctx context.Context, analytics *events.AnalyticsEvent) error {
	// Serialize analytics
	data, err := events.Marshal(analytics)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics: %w", err)
	}

	// Create Kafka message
	msg := kafka.Message{
		Key:   []byte(analytics.OrderID),
		Value: data,
		Time:  analytics.Timestamp,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(analytics.EventType)},
			{Key: "analytics-id", Value: []byte(analytics.ID)},
		},
	}

	// Get writer for analytics topic
	writer := p.getWriter("analytics")

	// Write message
	err = writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write analytics: %w", err)
	}

	log.Printf("Produced analytics event %s", analytics.ID)
	return nil
}

// Close closes all Kafka writers
func (p *Producer) Close() error {
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			log.Printf("Error closing writer for topic %s: %v", topic, err)
		}
	}
	return nil
}
