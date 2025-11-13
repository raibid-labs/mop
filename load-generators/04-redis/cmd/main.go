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

	"github.com/raibid-labs/mop/load-generators/04-redis/internal/generator"
	"github.com/raibid-labs/mop/load-generators/04-redis/internal/patterns"
)

var (
	redisAddr    = flag.String("redis-addr", getEnv("REDIS_ADDR", "localhost:6379"), "Redis address")
	redisPassword = flag.String("redis-password", getEnv("REDIS_PASSWORD", ""), "Redis password")
	redisDB      = flag.Int("redis-db", getIntEnv("REDIS_DB", 0), "Redis database")
	pattern      = flag.String("pattern", getEnv("LOAD_PATTERN", "constant"), "Load pattern: constant, spike, ramp")
	duration     = flag.Duration("duration", getDurationEnv("DURATION", 60*time.Second), "Test duration")
	rps          = flag.Int("rps", getIntEnv("RPS", 1000), "Operations per second")
	maxRPS       = flag.Int("max-rps", getIntEnv("MAX_RPS", 5000), "Maximum RPS")
	clients      = flag.Int("clients", getIntEnv("CLIENTS", 10), "Number of concurrent clients")
	opType       = flag.String("op-type", getEnv("OP_TYPE", "mixed"), "Operation type: get, set, mixed")
	keySize      = flag.Int("key-size", getIntEnv("KEY_SIZE", 16), "Key size in bytes")
	valueSize    = flag.Int("value-size", getIntEnv("VALUE_SIZE", 256), "Value size in bytes")
	metricsPort  = flag.Int("metrics-port", getIntEnv("METRICS_PORT", 9093), "Prometheus metrics port")
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
		loadPattern = patterns.NewConstantLoad(*rps)
	case "spike":
		loadPattern = patterns.NewSpikeLoad(*rps, *maxRPS, *duration/10)
	case "ramp":
		loadPattern = patterns.NewRampLoad(*rps, *maxRPS, *duration)
	default:
		log.Fatalf("Unknown load pattern: %s", *pattern)
	}

	gen, err := generator.New(generator.Config{
		RedisAddr:     *redisAddr,
		RedisPassword: *redisPassword,
		RedisDB:       *redisDB,
		Clients:       *clients,
		OpType:        *opType,
		KeySize:       *keySize,
		ValueSize:     *valueSize,
		MetricsPort:   *metricsPort,
	})
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}
	defer gen.Close()

	go gen.StartMetricsServer()

	log.Printf("Starting Redis load test against %s", *redisAddr)
	log.Printf("Pattern: %s, Duration: %s, Clients: %d, Op Type: %s", *pattern, *duration, *clients, *opType)

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
