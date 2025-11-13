package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/raibid-labs/mop/examples/05-kafka-streaming/internal/events"
	"github.com/raibid-labs/mop/examples/05-kafka-streaming/internal/handlers"
)

func main() {
	// Get Kafka brokers from environment
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092"
	}
	brokers := strings.Split(brokersEnv, ",")

	log.Printf("Starting Kafka producer, connecting to brokers: %v", brokers)

	// Create producer
	producer := handlers.NewProducer(brokers)
	defer producer.Close()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping producer...")
		cancel()
	}()

	// Generate events continuously
	rand.Seed(time.Now().UnixNano())
	eventCount := 0

	log.Println("Producer started, generating order events...")

	for {
		select {
		case <-ctx.Done():
			log.Printf("Producer stopped. Total events produced: %d", eventCount)
			return
		case <-time.After(2 * time.Second):
			// Generate random order event
			event := generateRandomOrderEvent()

			if err := producer.ProduceOrderEvent(ctx, event); err != nil {
				log.Printf("Error producing event: %v", err)
				continue
			}

			eventCount++

			// Occasionally generate updates and cancellations
			if eventCount%5 == 0 {
				updateEvent := generateUpdateEvent(event)
				if err := producer.ProduceOrderEvent(ctx, updateEvent); err != nil {
					log.Printf("Error producing update event: %v", err)
				}
				eventCount++
			}

			if eventCount%15 == 0 {
				cancelEvent := generateCancelEvent(event)
				if err := producer.ProduceOrderEvent(ctx, cancelEvent); err != nil {
					log.Printf("Error producing cancel event: %v", err)
				}
				eventCount++
			}
		}
	}
}

func generateRandomOrderEvent() *events.OrderEvent {
	orderID := fmt.Sprintf("ord-%d", rand.Intn(10000))
	customerID := fmt.Sprintf("cust-%d", rand.Intn(1000))
	eventID := fmt.Sprintf("evt-%d", time.Now().UnixNano())

	itemCount := rand.Intn(5) + 1
	items := make([]events.OrderItem, itemCount)
	total := 0.0

	for i := 0; i < itemCount; i++ {
		price := float64(rand.Intn(100)) + 9.99
		quantity := rand.Intn(3) + 1
		items[i] = events.OrderItem{
			ProductID: fmt.Sprintf("prod-%d", rand.Intn(100)),
			Name:      fmt.Sprintf("Product %d", rand.Intn(100)),
			Quantity:  quantity,
			Price:     price,
		}
		total += price * float64(quantity)
	}

	return &events.OrderEvent{
		ID:         eventID,
		Type:       events.OrderCreated,
		OrderID:    orderID,
		CustomerID: customerID,
		Status:     events.StatusPending,
		Total:      total,
		Items:      items,
		Timestamp:  time.Now(),
		Version:    1,
		Metadata: map[string]any{
			"source":   "producer-app",
			"region":   "us-west-2",
			"channel":  "web",
		},
	}
}

func generateUpdateEvent(original *events.OrderEvent) *events.OrderEvent {
	statuses := []events.OrderStatus{
		events.StatusConfirmed,
		events.StatusShipped,
		events.StatusDelivered,
	}

	return &events.OrderEvent{
		ID:         fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Type:       events.OrderUpdated,
		OrderID:    original.OrderID,
		CustomerID: original.CustomerID,
		Status:     statuses[rand.Intn(len(statuses))],
		Total:      original.Total,
		Items:      original.Items,
		Timestamp:  time.Now(),
		Version:    original.Version + 1,
		Metadata: map[string]any{
			"source": "producer-app",
			"reason": "status_change",
		},
	}
}

func generateCancelEvent(original *events.OrderEvent) *events.OrderEvent {
	return &events.OrderEvent{
		ID:         fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Type:       events.OrderCancelled,
		OrderID:    original.OrderID,
		CustomerID: original.CustomerID,
		Status:     events.StatusCancelled,
		Total:      original.Total,
		Items:      original.Items,
		Timestamp:  time.Now(),
		Version:    original.Version + 1,
		Metadata: map[string]any{
			"source": "producer-app",
			"reason": "customer_request",
		},
	}
}
