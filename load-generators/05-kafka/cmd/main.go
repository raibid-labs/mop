package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/raibid-labs/mop/load-generators/05-kafka/internal/generator"
	"github.com/raibid-labs/mop/load-generators/05-kafka/internal/patterns"
)

var (
	brokers      = flag.String("brokers", getEnv("KAFKA_BROKERS", "localhost:9092"), "Kafka broker addresses (comma-separated)")
	topic        = flag.String("topic", getEnv("KAFKA_TOPIC", "load-test"), "Kafka topic")
	pattern      = flag.String("pattern", getEnv("LOAD_PATTERN", "constant"), "Load pattern: constant, spike, ramp")
	duration     = flag.Duration("duration", getDurationEnv("DURATION", 60*time.Second), "Test duration")
	mps          = flag.Int("mps", getIntEnv("MPS", 100), "Messages per second")
	maxMPS       = flag.Int("max-mps", getIntEnv("MAX_MPS", 1000), "Maximum MPS")
	producers    = flag.Int("producers", getIntEnv("PRODUCERS", 3), "Number of concurrent producers")
	messageSize  = flag.Int("message-size", getIntEnv("MESSAGE_SIZE", 1024), "Message size in bytes")
	compression  = flag.String("compression", getEnv("COMPRESSION", "none"), "Compression: none, gzip, snappy, lz4")
	metricsPort  = flag.Int("metrics-port", getIntEnv("METRICS_PORT", 9094), "Prometheus metrics port")
	reportFormat = flag.String("report", getEnv("REPORT_FORMAT", "text"), "Report format: text, json")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	var loadPattern patterns.LoadPattern
	switch *pattern {
	case "constant":
		loadPattern = patterns.NewConstantLoad(*mps)
	case "spike":
		loadPattern = patterns.NewSpikeLoad(*mps, *maxMPS, *duration/10)
	case "ramp":
		loadPattern = patterns.NewRampLoad(*mps, *maxMPS, *duration)
	default:
		log.Fatalf("Unknown load pattern: %s", *pattern)
	}

	brokerList := strings.Split(*brokers, ",")

	gen, err := generator.New(generator.Config{
		Brokers:     brokerList,
		Topic:       *topic,
		Producers:   *producers,
		MessageSize: *messageSize,
		Compression: *compression,
		MetricsPort: *metricsPort,
	})
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}
	defer gen.Close()

	go gen.StartMetricsServer()

	log.Printf("Starting Kafka load test against %s", *brokers)
	log.Printf("Topic: %s, Pattern: %s, Duration: %s, Producers: %d", *topic, *pattern, *duration, *producers)

	ctx, timeoutCancel := context.WithTimeout(ctx, *duration)
	defer timeoutCancel()

	results := gen.Run(ctx, loadPattern)

	if *reportFormat == "json" {
		fmt.Println(results.ToJSON())
	} else {
		fmt.Println(results.ToString())
	}

	if results.FailureRate() > 0.05 {
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
