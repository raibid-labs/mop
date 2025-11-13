package generator

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/raibid-labs/mop/load-generators/04-redis/internal/patterns"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Clients       int
	OpType        string
	KeySize       int
	ValueSize     int
	MetricsPort   int
}

type Generator struct {
	config  Config
	client  *redis.Client
	metrics *Metrics
}

type Results struct {
	TotalOperations   int64
	SuccessOperations int64
	FailedOperations  int64
	TotalDuration     time.Duration
	MinLatency        time.Duration
	MaxLatency        time.Duration
	AvgLatency        time.Duration
	P50Latency        time.Duration
	P95Latency        time.Duration
	P99Latency        time.Duration
	Latencies         []time.Duration
	Errors            map[string]int64
}

type Metrics struct {
	operationsTotal  *prometheus.CounterVec
	operationDuration *prometheus.HistogramVec
	activeConnections prometheus.Gauge
}

func New(config Config) (*Generator, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
		PoolSize: config.Clients * 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	metrics := &Metrics{
		operationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_load_operations_total",
				Help: "Total number of Redis operations",
			},
			[]string{"status", "type"},
		),
		operationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "redis_load_operation_duration_seconds",
				Help:    "Redis operation duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status", "type"},
		),
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "redis_load_active_connections",
				Help: "Number of active Redis connections",
			},
		),
	}

	prometheus.MustRegister(metrics.operationsTotal)
	prometheus.MustRegister(metrics.operationDuration)
	prometheus.MustRegister(metrics.activeConnections)

	return &Generator{
		config:  config,
		client:  client,
		metrics: metrics,
	}, nil
}

func (g *Generator) Close() {
	if g.client != nil {
		g.client.Close()
	}
}

func (g *Generator) Run(ctx context.Context, pattern patterns.LoadPattern) *Results {
	results := &Results{
		Errors:    make(map[string]int64),
		Latencies: make([]time.Duration, 0),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	startTime := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	opChan := make(chan struct{}, g.config.Clients*10)

	for i := 0; i < g.config.Clients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.worker(ctx, opChan, results, &mu)
		}()
	}

	go func() {
		defer close(opChan)
		currentSecond := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rps := pattern.RPS(currentSecond)
				for i := 0; i < rps; i++ {
					select {
					case opChan <- struct{}{}:
					case <-ctx.Done():
						return
					}
				}
				currentSecond++
			}
		}
	}()

	wg.Wait()
	results.TotalDuration = time.Since(startTime)
	g.calculateStatistics(results)
	return results
}

func (g *Generator) worker(ctx context.Context, opChan <-chan struct{}, results *Results, mu *sync.Mutex) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-opChan:
			if !ok {
				return
			}
			g.executeOperation(ctx, results, mu)
		}
	}
}

func (g *Generator) executeOperation(ctx context.Context, results *Results, mu *sync.Mutex) {
	g.metrics.activeConnections.Inc()
	defer g.metrics.activeConnections.Dec()

	atomic.AddInt64(&results.TotalOperations, 1)

	opType := g.config.OpType
	if opType == "mixed" {
		if rand.Float64() < 0.8 {
			opType = "get"
		} else {
			opType = "set"
		}
	}

	start := time.Now()
	var err error

	key := fmt.Sprintf("loadtest:key:%d", rand.Intn(10000))

	switch opType {
	case "get":
		_, err = g.client.Get(ctx, key).Result()
		if err == redis.Nil {
			err = nil // Key not found is not an error for load testing
		}
	case "set":
		value := make([]byte, g.config.ValueSize)
		rand.Read(value)
		err = g.client.Set(ctx, key, value, time.Hour).Err()
	}

	latency := time.Since(start)

	mu.Lock()
	results.Latencies = append(results.Latencies, latency)
	mu.Unlock()

	if err != nil {
		atomic.AddInt64(&results.FailedOperations, 1)
		mu.Lock()
		results.Errors[err.Error()]++
		mu.Unlock()
		g.metrics.operationsTotal.WithLabelValues("error", opType).Inc()
		g.metrics.operationDuration.WithLabelValues("error", opType).Observe(latency.Seconds())
		return
	}

	atomic.AddInt64(&results.SuccessOperations, 1)
	g.metrics.operationsTotal.WithLabelValues("success", opType).Inc()
	g.metrics.operationDuration.WithLabelValues("success", opType).Observe(latency.Seconds())
}

func (g *Generator) calculateStatistics(results *Results) {
	if len(results.Latencies) == 0 {
		return
	}

	latencies := results.Latencies
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}

	results.MinLatency = latencies[0]
	results.MaxLatency = latencies[len(latencies)-1]

	var sum time.Duration
	for _, l := range latencies {
		sum += l
	}
	results.AvgLatency = sum / time.Duration(len(latencies))

	results.P50Latency = latencies[len(latencies)*50/100]
	results.P95Latency = latencies[len(latencies)*95/100]
	results.P99Latency = latencies[len(latencies)*99/100]
}

func (g *Generator) StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	addr := fmt.Sprintf(":%d", g.config.MetricsPort)
	http.ListenAndServe(addr, nil)
}

func (r *Results) FailureRate() float64 {
	if r.TotalOperations == 0 {
		return 0
	}
	return float64(r.FailedOperations) / float64(r.TotalOperations)
}

func (r *Results) ToString() string {
	return fmt.Sprintf(`
=== Redis Load Test Results ===
Total Operations:   %d
Success:            %d (%.2f%%)
Failed:             %d (%.2f%%)
Duration:           %s
OPS:                %.2f

Latency Statistics:
  Min:              %s
  Max:              %s
  Avg:              %s
  P50:              %s
  P95:              %s
  P99:              %s
`, r.TotalOperations, r.SuccessOperations,
		float64(r.SuccessOperations)/float64(r.TotalOperations)*100,
		r.FailedOperations, r.FailureRate()*100,
		r.TotalDuration,
		float64(r.TotalOperations)/r.TotalDuration.Seconds(),
		r.MinLatency, r.MaxLatency, r.AvgLatency,
		r.P50Latency, r.P95Latency, r.P99Latency)
}

func (r *Results) ToJSON() string {
	return fmt.Sprintf(`{
  "total_operations": %d,
  "success_operations": %d,
  "failed_operations": %d,
  "duration_seconds": %.2f,
  "ops": %.2f,
  "latency": {
    "min_ms": %.2f,
    "max_ms": %.2f,
    "avg_ms": %.2f,
    "p50_ms": %.2f,
    "p95_ms": %.2f,
    "p99_ms": %.2f
  }
}`, r.TotalOperations, r.SuccessOperations, r.FailedOperations,
		r.TotalDuration.Seconds(),
		float64(r.TotalOperations)/r.TotalDuration.Seconds(),
		float64(r.MinLatency.Microseconds())/1000,
		float64(r.MaxLatency.Microseconds())/1000,
		float64(r.AvgLatency.Microseconds())/1000,
		float64(r.P50Latency.Microseconds())/1000,
		float64(r.P95Latency.Microseconds())/1000,
		float64(r.P99Latency.Microseconds())/1000)
}
