package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/raibid-labs/mop/load-generators/02-grpc/internal/patterns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Config struct {
	Target      string
	Method      string
	Data        string
	Concurrency int
	Timeout     time.Duration
	MetricsPort int
	Insecure    bool
}

type Generator struct {
	config  Config
	conn    *grpc.ClientConn
	metrics *Metrics
}

type Results struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalDuration   time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	AvgLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	Latencies       []time.Duration
	StatusCodes     map[codes.Code]int64
	Errors          map[string]int64
}

type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestsInFlight prometheus.Gauge
}

func New(config Config) (*Generator, error) {
	// Setup gRPC connection
	var opts []grpc.DialOption
	if config.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(config.Target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", config.Target, err)
	}

	metrics := &Metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_load_requests_total",
				Help: "Total number of gRPC requests",
			},
			[]string{"status", "method"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "grpc_load_request_duration_seconds",
				Help:    "gRPC request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status", "method"},
		),
		requestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "grpc_load_requests_in_flight",
				Help: "Number of gRPC requests currently in flight",
			},
		),
	}

	prometheus.MustRegister(metrics.requestsTotal)
	prometheus.MustRegister(metrics.requestDuration)
	prometheus.MustRegister(metrics.requestsInFlight)

	return &Generator{
		config:  config,
		conn:    conn,
		metrics: metrics,
	}, nil
}

func (g *Generator) Close() {
	if g.conn != nil {
		g.conn.Close()
	}
}

func (g *Generator) Run(ctx context.Context, pattern patterns.LoadPattern) *Results {
	results := &Results{
		StatusCodes: make(map[codes.Code]int64),
		Errors:      make(map[string]int64),
		Latencies:   make([]time.Duration, 0),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	startTime := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	reqChan := make(chan struct{}, g.config.Concurrency*10)

	// Start workers
	for i := 0; i < g.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.worker(ctx, reqChan, results, &mu)
		}()
	}

	// Generate load according to pattern
	go func() {
		defer close(reqChan)
		currentSecond := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rps := pattern.RPS(currentSecond)
				for i := 0; i < rps; i++ {
					select {
					case reqChan <- struct{}{}:
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

	// Calculate statistics
	g.calculateStatistics(results)

	return results
}

func (g *Generator) worker(ctx context.Context, reqChan <-chan struct{}, results *Results, mu *sync.Mutex) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-reqChan:
			if !ok {
				return
			}
			g.makeRequest(ctx, results, mu)
		}
	}
}

func (g *Generator) makeRequest(ctx context.Context, results *Results, mu *sync.Mutex) {
	g.metrics.requestsInFlight.Inc()
	defer g.metrics.requestsInFlight.Dec()

	atomic.AddInt64(&results.TotalRequests, 1)

	// Parse request data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(g.config.Data), &data); err != nil {
		atomic.AddInt64(&results.FailedRequests, 1)
		mu.Lock()
		results.Errors[err.Error()]++
		mu.Unlock()
		return
	}

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, g.config.Timeout)
	defer cancel()

	start := time.Now()

	// Make gRPC call using dynamic stub
	err := g.conn.Invoke(reqCtx, "/"+g.config.Method, data, &data)

	latency := time.Since(start)

	mu.Lock()
	results.Latencies = append(results.Latencies, latency)
	mu.Unlock()

	if err != nil {
		atomic.AddInt64(&results.FailedRequests, 1)

		// Extract gRPC status code
		st, ok := status.FromError(err)
		if ok {
			mu.Lock()
			results.StatusCodes[st.Code()]++
			mu.Unlock()
			g.metrics.requestsTotal.WithLabelValues(st.Code().String(), g.config.Method).Inc()
			g.metrics.requestDuration.WithLabelValues(st.Code().String(), g.config.Method).Observe(latency.Seconds())
		} else {
			mu.Lock()
			results.Errors[err.Error()]++
			mu.Unlock()
			g.metrics.requestsTotal.WithLabelValues("error", g.config.Method).Inc()
			g.metrics.requestDuration.WithLabelValues("error", g.config.Method).Observe(latency.Seconds())
		}
		return
	}

	atomic.AddInt64(&results.SuccessRequests, 1)

	mu.Lock()
	results.StatusCodes[codes.OK]++
	mu.Unlock()

	g.metrics.requestsTotal.WithLabelValues(codes.OK.String(), g.config.Method).Inc()
	g.metrics.requestDuration.WithLabelValues(codes.OK.String(), g.config.Method).Observe(latency.Seconds())
}

func (g *Generator) calculateStatistics(results *Results) {
	if len(results.Latencies) == 0 {
		return
	}

	// Sort latencies
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
	if r.TotalRequests == 0 {
		return 0
	}
	return float64(r.FailedRequests) / float64(r.TotalRequests)
}

func (r *Results) ToString() string {
	var sb string
	sb += "\n=== gRPC Load Test Results ===\n"
	sb += fmt.Sprintf("Total Requests:   %d\n", r.TotalRequests)
	sb += fmt.Sprintf("Success:          %d (%.2f%%)\n", r.SuccessRequests, float64(r.SuccessRequests)/float64(r.TotalRequests)*100)
	sb += fmt.Sprintf("Failed:           %d (%.2f%%)\n", r.FailedRequests, r.FailureRate()*100)
	sb += fmt.Sprintf("Duration:         %s\n", r.TotalDuration)
	sb += fmt.Sprintf("RPS:              %.2f\n", float64(r.TotalRequests)/r.TotalDuration.Seconds())
	sb += "\nLatency Statistics:\n"
	sb += fmt.Sprintf("  Min:            %s\n", r.MinLatency)
	sb += fmt.Sprintf("  Max:            %s\n", r.MaxLatency)
	sb += fmt.Sprintf("  Avg:            %s\n", r.AvgLatency)
	sb += fmt.Sprintf("  P50:            %s\n", r.P50Latency)
	sb += fmt.Sprintf("  P95:            %s\n", r.P95Latency)
	sb += fmt.Sprintf("  P99:            %s\n", r.P99Latency)
	sb += "\nStatus Codes:\n"
	for code, count := range r.StatusCodes {
		sb += fmt.Sprintf("  %s:             %d\n", code.String(), count)
	}
	if len(r.Errors) > 0 {
		sb += "\nErrors:\n"
		for err, count := range r.Errors {
			sb += fmt.Sprintf("  %s: %d\n", err, count)
		}
	}
	return sb
}

func (r *Results) ToJSON() string {
	return fmt.Sprintf(`{
  "total_requests": %d,
  "success_requests": %d,
  "failed_requests": %d,
  "duration_seconds": %.2f,
  "rps": %.2f,
  "latency": {
    "min_ms": %.2f,
    "max_ms": %.2f,
    "avg_ms": %.2f,
    "p50_ms": %.2f,
    "p95_ms": %.2f,
    "p99_ms": %.2f
  }
}`, r.TotalRequests, r.SuccessRequests, r.FailedRequests,
		r.TotalDuration.Seconds(),
		float64(r.TotalRequests)/r.TotalDuration.Seconds(),
		float64(r.MinLatency.Microseconds())/1000,
		float64(r.MaxLatency.Microseconds())/1000,
		float64(r.AvgLatency.Microseconds())/1000,
		float64(r.P50Latency.Microseconds())/1000,
		float64(r.P95Latency.Microseconds())/1000,
		float64(r.P99Latency.Microseconds())/1000)
}
