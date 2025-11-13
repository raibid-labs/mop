package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/raibid-labs/mop/examples/05-kafka-streaming/internal/handlers"
)

func main() {
	// Get configuration from environment
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092"
	}
	brokers := strings.Split(brokersEnv, ",")

	consumerType := os.Getenv("CONSUMER_TYPE")
	if consumerType == "" {
		consumerType = "notifications"
	}

	groupID := os.Getenv("CONSUMER_GROUP")
	if groupID == "" {
		groupID = fmt.Sprintf("%s-group", consumerType)
	}

	log.Printf("Starting Kafka consumer (type: %s, group: %s), connecting to brokers: %v",
		consumerType, groupID, brokers)

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping consumer...")
		cancel()
	}()

	var consumer *handlers.Consumer

	switch consumerType {
	case "notifications":
		cfg := handlers.ConsumerConfig{
			Brokers:    brokers,
			Topic:      "notifications",
			GroupID:    groupID,
			MaxRetries: 3,
			DLQEnabled: true,
		}
		consumer = handlers.NewConsumer(cfg, handlers.NotificationHandler)

	case "analytics":
		cfg := handlers.ConsumerConfig{
			Brokers:    brokers,
			Topic:      "analytics",
			GroupID:    groupID,
			MaxRetries: 3,
			DLQEnabled: true,
		}
		consumer = handlers.NewConsumer(cfg, handlers.AnalyticsHandler)

	case "orders":
		// Create a producer for downstream events
		producer := handlers.NewProducer(brokers)
		defer producer.Close()

		// Subscribe to all order topics (we'll need to handle multiple topics)
		// For simplicity, we'll use orders.created
		cfg := handlers.ConsumerConfig{
			Brokers:    brokers,
			Topic:      "orders.created",
			GroupID:    groupID,
			MaxRetries: 3,
			DLQEnabled: true,
		}
		consumer = handlers.NewConsumer(cfg, handlers.OrderEventHandler(producer))

	default:
		log.Fatalf("Unknown consumer type: %s", consumerType)
	}

	log.Printf("Consumer started (type: %s)", consumerType)

	// Start consuming
	if err := consumer.Start(ctx); err != nil {
		log.Printf("Consumer stopped: %v", err)
	}

	log.Printf("Consumer shutdown complete")
}
