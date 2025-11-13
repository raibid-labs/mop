package generator

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/raibid-labs/mop/load-generators/01-http/internal/patterns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	TargetURL   string
	Method      string
	Body        string
	Headers     map[string]string
	Concurrency int
	Timeout     time.Duration
	MetricsPort int
}

type Generator struct {
	config  Config
	client  *http.Client
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
	StatusCodes     map[int]int64
	Errors          map[string]int64
}

type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestsInFlight prometheus.Gauge
}

func New(config Config) *Generator {
	metrics := &Metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_load_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"status", "method"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_load_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status", "method"},
		),
		requestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_load_requests_in_flight",
				Help: "Number of HTTP requests currently in flight",
			},
		),
	}

	prometheus.MustRegister(metrics.requestsTotal)
	prometheus.MustRegister(metrics.requestDuration)
	prometheus.MustRegister(metrics.requestsInFlight)

	return &Generator{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        config.Concurrency * 2,
				MaxIdleConnsPerHost: config.Concurrency * 2,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		metrics: metrics,
	}
}

func (g *Generator) Run(ctx context.Context, pattern patterns.LoadPattern) *Results {
	results := &Results{
		StatusCodes: make(map[int]int64),
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

	var body io.Reader
	if g.config.Body != "" {
		body = strings.NewReader(g.config.Body)
	}

	req, err := http.NewRequestWithContext(ctx, g.config.Method, g.config.TargetURL, body)
	if err != nil {
		atomic.AddInt64(&results.FailedRequests, 1)
		mu.Lock()
		results.Errors[err.Error()]++
		mu.Unlock()
		return
	}

	for key, value := range g.config.Headers {
		req.Header.Set(key, value)
	}

	start := time.Now()
	resp, err := g.client.Do(req)
	latency := time.Since(start)

	mu.Lock()
	results.Latencies = append(results.Latencies, latency)
	mu.Unlock()

	if err != nil {
		atomic.AddInt64(&results.FailedRequests, 1)
		mu.Lock()
		results.Errors[err.Error()]++
		mu.Unlock()
		g.metrics.requestsTotal.WithLabelValues("error", g.config.Method).Inc()
		g.metrics.requestDuration.WithLabelValues("error", g.config.Method).Observe(latency.Seconds())
		return
	}
	defer resp.Body.Close()

	// Drain response body
	io.Copy(io.Discard, resp.Body)

	mu.Lock()
	results.StatusCodes[resp.StatusCode]++
	mu.Unlock()

	status := fmt.Sprintf("%d", resp.StatusCode)
	g.metrics.requestsTotal.WithLabelValues(status, g.config.Method).Inc()
	g.metrics.requestDuration.WithLabelValues(status, g.config.Method).Observe(latency.Seconds())

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		atomic.AddInt64(&results.SuccessRequests, 1)
	} else {
		atomic.AddInt64(&results.FailedRequests, 1)
	}
}

func (g *Generator) calculateStatistics(results *Results) {
	if len(results.Latencies) == 0 {
		return
	}

	// Sort latencies
	// Simple bubble sort for small datasets
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
	var sb strings.Builder
	sb.WriteString("\n=== Load Test Results ===\n")
	sb.WriteString(fmt.Sprintf("Total Requests:   %d\n", r.TotalRequests))
	sb.WriteString(fmt.Sprintf("Success:          %d (%.2f%%)\n", r.SuccessRequests, float64(r.SuccessRequests)/float64(r.TotalRequests)*100))
	sb.WriteString(fmt.Sprintf("Failed:           %d (%.2f%%)\n", r.FailedRequests, r.FailureRate()*100))
	sb.WriteString(fmt.Sprintf("Duration:         %s\n", r.TotalDuration))
	sb.WriteString(fmt.Sprintf("RPS:              %.2f\n", float64(r.TotalRequests)/r.TotalDuration.Seconds()))
	sb.WriteString("\nLatency Statistics:\n")
	sb.WriteString(fmt.Sprintf("  Min:            %s\n", r.MinLatency))
	sb.WriteString(fmt.Sprintf("  Max:            %s\n", r.MaxLatency))
	sb.WriteString(fmt.Sprintf("  Avg:            %s\n", r.AvgLatency))
	sb.WriteString(fmt.Sprintf("  P50:            %s\n", r.P50Latency))
	sb.WriteString(fmt.Sprintf("  P95:            %s\n", r.P95Latency))
	sb.WriteString(fmt.Sprintf("  P99:            %s\n", r.P99Latency))
	sb.WriteString("\nStatus Codes:\n")
	for code, count := range r.StatusCodes {
		sb.WriteString(fmt.Sprintf("  %d:             %d\n", code, count))
	}
	if len(r.Errors) > 0 {
		sb.WriteString("\nErrors:\n")
		for err, count := range r.Errors {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", err, count))
		}
	}
	return sb.String()
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
