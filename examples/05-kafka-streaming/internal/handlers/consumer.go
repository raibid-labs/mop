package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/raibid-labs/mop/examples/05-kafka-streaming/internal/events"
	"github.com/segmentio/kafka-go"
)

// MessageHandler is a function that processes a message
type MessageHandler func(ctx context.Context, msg kafka.Message) error

// Consumer wraps a Kafka reader for consuming messages
type Consumer struct {
	reader      *kafka.Reader
	dlqProducer *Producer
	handler     MessageHandler
	maxRetries  int
}

// ConsumerConfig contains configuration for a consumer
type ConsumerConfig struct {
	Brokers      []string
	Topic        string
	GroupID      string
	MaxRetries   int
	DLQEnabled   bool
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg ConsumerConfig, handler MessageHandler) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
		MaxWait:        500 * time.Millisecond,
	})

	consumer := &Consumer{
		reader:     reader,
		handler:    handler,
		maxRetries: cfg.MaxRetries,
	}

	// Set up DLQ producer if enabled
	if cfg.DLQEnabled {
		consumer.dlqProducer = NewProducer(cfg.Brokers)
	}

	return consumer
}

// Start begins consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("Starting consumer for topic %s in group %s", c.reader.Config().Topic, c.reader.Config().GroupID)

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping consumer")
			return c.Close()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					// Context was cancelled
					return c.Close()
				}
				log.Printf("Error fetching message: %v", err)
				continue
			}

			// Process message with retries
			if err := c.processMessageWithRetry(ctx, msg); err != nil {
				log.Printf("Failed to process message after retries: %v", err)

				// Send to DLQ if enabled
				if c.dlqProducer != nil {
					if dlqErr := c.sendToDLQ(ctx, msg, err); dlqErr != nil {
						log.Printf("Failed to send message to DLQ: %v", dlqErr)
					}
				}
			}

			// Commit message
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Failed to commit message: %v", err)
			}
		}
	}
}

// processMessageWithRetry processes a message with retry logic
func (c *Consumer) processMessageWithRetry(ctx context.Context, msg kafka.Message) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
			log.Printf("Retrying message (attempt %d/%d) after %v", attempt, c.maxRetries, backoff)
			time.Sleep(backoff)
		}

		err := c.handler(ctx, msg)
		if err == nil {
			if attempt > 0 {
				log.Printf("Message processed successfully after %d retries", attempt)
			}
			return nil
		}

		lastErr = err
		log.Printf("Error processing message (attempt %d/%d): %v", attempt+1, c.maxRetries+1, err)
	}

	return fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
}

// sendToDLQ sends a failed message to the dead letter queue
func (c *Consumer) sendToDLQ(ctx context.Context, msg kafka.Message, processingErr error) error {
	dlqMsg := kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
		Time:  time.Now(),
		Headers: append(msg.Headers,
			kafka.Header{Key: "dlq-reason", Value: []byte(processingErr.Error())},
			kafka.Header{Key: "original-topic", Value: []byte(c.reader.Config().Topic)},
			kafka.Header{Key: "original-partition", Value: []byte(fmt.Sprintf("%d", msg.Partition))},
			kafka.Header{Key: "original-offset", Value: []byte(fmt.Sprintf("%d", msg.Offset))},
			kafka.Header{Key: "failed-at", Value: []byte(time.Now().Format(time.RFC3339))},
		),
	}

	writer := c.dlqProducer.getWriter("dlq")
	if err := writer.WriteMessages(ctx, dlqMsg); err != nil {
		return fmt.Errorf("failed to write to DLQ: %w", err)
	}

	log.Printf("Sent message to DLQ: key=%s, reason=%s", string(msg.Key), processingErr.Error())
	return nil
}

// Close closes the consumer and its resources
func (c *Consumer) Close() error {
	if c.dlqProducer != nil {
		if err := c.dlqProducer.Close(); err != nil {
			log.Printf("Error closing DLQ producer: %v", err)
		}
	}

	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}

	return nil
}

// NotificationHandler processes notification messages
func NotificationHandler(ctx context.Context, msg kafka.Message) error {
	notification, err := events.UnmarshalNotificationEvent(msg.Value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	log.Printf("Processing notification %s: %s to %s", notification.ID, notification.Message, notification.Email)

	// Simulate sending email
	time.Sleep(50 * time.Millisecond)

	// In a real implementation, you would send the actual email here
	log.Printf("Email sent to %s for order %s", notification.Email, notification.OrderID)

	return nil
}

// AnalyticsHandler processes analytics messages
func AnalyticsHandler(ctx context.Context, msg kafka.Message) error {
	analytics, err := events.UnmarshalAnalyticsEvent(msg.Value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal analytics: %w", err)
	}

	log.Printf("Processing analytics event %s: %s for order %s (total: $%.2f)",
		analytics.ID, analytics.EventType, analytics.OrderID, analytics.Total)

	// Simulate analytics processing
	time.Sleep(30 * time.Millisecond)

	// In a real implementation, you would write to an analytics database
	log.Printf("Analytics recorded: %s - %d items, status: %s",
		analytics.OrderID, analytics.ItemCount, analytics.Status)

	return nil
}

// OrderEventHandler processes order events and generates downstream events
func OrderEventHandler(producer *Producer) MessageHandler {
	return func(ctx context.Context, msg kafka.Message) error {
		orderEvent, err := events.UnmarshalOrderEvent(msg.Value)
		if err != nil {
			return fmt.Errorf("failed to unmarshal order event: %w", err)
		}

		log.Printf("Processing order event %s: %s for order %s",
			orderEvent.ID, orderEvent.Type, orderEvent.OrderID)

		// Generate notification event
		notification := &events.NotificationEvent{
			ID:         fmt.Sprintf("notif-%s", orderEvent.ID),
			OrderID:    orderEvent.OrderID,
			CustomerID: orderEvent.CustomerID,
			Type:       "email",
			Email:      fmt.Sprintf("customer-%s@example.com", orderEvent.CustomerID),
			Timestamp:  time.Now(),
		}

		switch orderEvent.Type {
		case events.OrderCreated:
			notification.Message = fmt.Sprintf("Your order %s has been created with total $%.2f",
				orderEvent.OrderID, orderEvent.Total)
		case events.OrderUpdated:
			notification.Message = fmt.Sprintf("Your order %s has been updated (status: %s)",
				orderEvent.OrderID, orderEvent.Status)
		case events.OrderCancelled:
			notification.Message = fmt.Sprintf("Your order %s has been cancelled",
				orderEvent.OrderID)
		}

		if err := producer.ProduceNotification(ctx, notification); err != nil {
			return fmt.Errorf("failed to produce notification: %w", err)
		}

		// Generate analytics event
		analytics := &events.AnalyticsEvent{
			ID:         fmt.Sprintf("analytics-%s", orderEvent.ID),
			EventType:  string(orderEvent.Type),
			OrderID:    orderEvent.OrderID,
			CustomerID: orderEvent.CustomerID,
			Total:      orderEvent.Total,
			ItemCount:  len(orderEvent.Items),
			Status:     orderEvent.Status,
			Timestamp:  time.Now(),
			Metadata:   orderEvent.Metadata,
		}

		if err := producer.ProduceAnalytics(ctx, analytics); err != nil {
			return fmt.Errorf("failed to produce analytics: %w", err)
		}

		log.Printf("Generated downstream events for order %s", orderEvent.OrderID)
		return nil
	}
}
