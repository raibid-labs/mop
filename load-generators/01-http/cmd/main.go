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

	"github.com/raibid-labs/mop/load-generators/01-http/internal/generator"
	"github.com/raibid-labs/mop/load-generators/01-http/internal/patterns"
)

var (
	targetURL    = flag.String("target", getEnv("TARGET_URL", "http://localhost:8080"), "Target URL to load test")
	pattern      = flag.String("pattern", getEnv("LOAD_PATTERN", "constant"), "Load pattern: constant, spike, ramp")
	duration     = flag.Duration("duration", getDurationEnv("DURATION", 60*time.Second), "Test duration")
	rps          = flag.Int("rps", getIntEnv("RPS", 100), "Requests per second for constant load")
	maxRPS       = flag.Int("max-rps", getIntEnv("MAX_RPS", 500), "Maximum RPS for spike/ramp patterns")
	concurrency  = flag.Int("concurrency", getIntEnv("CONCURRENCY", 10), "Number of concurrent workers")
	timeout      = flag.Duration("timeout", getDurationEnv("TIMEOUT", 30*time.Second), "Request timeout")
	method       = flag.String("method", getEnv("METHOD", "GET"), "HTTP method")
	body         = flag.String("body", getEnv("BODY", ""), "Request body")
	headers      = flag.String("headers", getEnv("HEADERS", ""), "Headers in key:value,key:value format")
	metricsPort  = flag.Int("metrics-port", getIntEnv("METRICS_PORT", 9090), "Prometheus metrics port")
	reportFormat = flag.String("report", getEnv("REPORT_FORMAT", "text"), "Report format: text, json")
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
	gen := generator.New(generator.Config{
		TargetURL:   *targetURL,
		Method:      *method,
		Body:        *body,
		Headers:     parseHeaders(*headers),
		Concurrency: *concurrency,
		Timeout:     *timeout,
		MetricsPort: *metricsPort,
	})

	// Start metrics server
	go gen.StartMetricsServer()

	log.Printf("Starting load test against %s", *targetURL)
	log.Printf("Pattern: %s, Duration: %s, Concurrency: %d", *pattern, *duration, *concurrency)

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

func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}

	// Parse format: "key1:value1,key2:value2"
	// Simple implementation, could be enhanced
	return headers
}
