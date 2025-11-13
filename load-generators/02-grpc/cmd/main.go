package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raibid-labs/mop/load-generators/02-grpc/internal/generator"
	"github.com/raibid-labs/mop/load-generators/02-grpc/internal/patterns"
)

var (
	target       = flag.String("target", getEnv("TARGET", "localhost:9090"), "gRPC server address")
	method       = flag.String("method", getEnv("METHOD", "auth.v1.AuthService/Login"), "gRPC method to call")
	pattern      = flag.String("pattern", getEnv("LOAD_PATTERN", "constant"), "Load pattern: constant, spike, ramp")
	duration     = flag.Duration("duration", getDurationEnv("DURATION", 60*time.Second), "Test duration")
	rps          = flag.Int("rps", getIntEnv("RPS", 100), "Requests per second for constant load")
	maxRPS       = flag.Int("max-rps", getIntEnv("MAX_RPS", 500), "Maximum RPS for spike/ramp patterns")
	concurrency  = flag.Int("concurrency", getIntEnv("CONCURRENCY", 10), "Number of concurrent workers")
	timeout      = flag.Duration("timeout", getDurationEnv("TIMEOUT", 30*time.Second), "Request timeout")
	data         = flag.String("data", getEnv("DATA", `{"username":"loadtest","password":"password"}`), "Request data as JSON")
	metricsPort  = flag.Int("metrics-port", getIntEnv("METRICS_PORT", 9091), "Prometheus metrics port")
	reportFormat = flag.String("report", getEnv("REPORT_FORMAT", "text"), "Report format: text, json")
	insecure     = flag.Bool("insecure", getBoolEnv("INSECURE", true), "Use insecure connection")
)

func main() {
	flag.Parse()

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	// Create load pattern
	var loadPattern patterns.LoadPattern
	switch *pattern {
	case "constant":
		loadPattern = patterns.NewConstantLoad(*rps)
	case "spike":
		loadPattern = patterns.NewSpikeLoad(*rps, *maxRPS, *duration/10)
	case "ramp":
		loadPattern = patterns.NewRampLoad(*rps, *maxRPS, *duration)
	default:
		log.Fatalf("Unknown load pattern: %s", *pattern)
	}

	// Create and configure generator
	gen, err := generator.New(generator.Config{
		Target:      *target,
		Method:      *method,
		Data:        *data,
		Concurrency: *concurrency,
		Timeout:     *timeout,
		MetricsPort: *metricsPort,
		Insecure:    *insecure,
	})
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}
	defer gen.Close()

	// Start metrics server
	go gen.StartMetricsServer()

	log.Printf("Starting gRPC load test against %s", *target)
	log.Printf("Method: %s, Pattern: %s, Duration: %s", *method, *pattern, *duration)

	// Run load test
	ctx, timeoutCancel := context.WithTimeout(ctx, *duration)
	defer timeoutCancel()

	results := gen.Run(ctx, loadPattern)

	// Print results
	if *reportFormat == "json" {
		fmt.Println(results.ToJSON())
	} else {
		fmt.Println(results.ToString())
	}

	// Exit with error code if too many failures
	if results.FailureRate() > 0.05 { // More than 5% failures
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

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
