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

	"github.com/raibid-labs/mop/load-generators/03-sql/internal/generator"
	"github.com/raibid-labs/mop/load-generators/03-sql/internal/patterns"
)

var (
	dbHost       = flag.String("db-host", getEnv("DB_HOST", "localhost"), "Database host")
	dbPort       = flag.Int("db-port", getIntEnv("DB_PORT", 5432), "Database port")
	dbUser       = flag.String("db-user", getEnv("DB_USER", "postgres"), "Database user")
	dbPassword   = flag.String("db-password", getEnv("DB_PASSWORD", "postgres"), "Database password")
	dbName       = flag.String("db-name", getEnv("DB_NAME", "orders"), "Database name")
	pattern      = flag.String("pattern", getEnv("LOAD_PATTERN", "constant"), "Load pattern: constant, spike, ramp")
	duration     = flag.Duration("duration", getDurationEnv("DURATION", 60*time.Second), "Test duration")
	tps          = flag.Int("tps", getIntEnv("TPS", 100), "Transactions per second for constant load")
	maxTPS       = flag.Int("max-tps", getIntEnv("MAX_TPS", 500), "Maximum TPS for spike/ramp patterns")
	clients      = flag.Int("clients", getIntEnv("CLIENTS", 10), "Number of concurrent clients")
	queryType    = flag.String("query-type", getEnv("QUERY_TYPE", "mixed"), "Query type: read, write, mixed")
	metricsPort  = flag.Int("metrics-port", getIntEnv("METRICS_PORT", 9092), "Prometheus metrics port")
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
		loadPattern = patterns.NewConstantLoad(*tps)
	case "spike":
		loadPattern = patterns.NewSpikeLoad(*tps, *maxTPS, *duration/10)
	case "ramp":
		loadPattern = patterns.NewRampLoad(*tps, *maxTPS, *duration)
	default:
		log.Fatalf("Unknown load pattern: %s", *pattern)
	}

	// Create and configure generator
	gen, err := generator.New(generator.Config{
		DBHost:      *dbHost,
		DBPort:      *dbPort,
		DBUser:      *dbUser,
		DBPassword:  *dbPassword,
		DBName:      *dbName,
		Clients:     *clients,
		QueryType:   *queryType,
		MetricsPort: *metricsPort,
	})
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}
	defer gen.Close()

	// Initialize database schema
	if err := gen.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start metrics server
	go gen.StartMetricsServer()

	log.Printf("Starting SQL load test against %s:%d/%s", *dbHost, *dbPort, *dbName)
	log.Printf("Pattern: %s, Duration: %s, Clients: %d, Query Type: %s", *pattern, *duration, *clients, *queryType)

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
