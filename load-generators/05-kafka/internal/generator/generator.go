package generator

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/raibid-labs/mop/load-generators/05-kafka/internal/patterns"
	"github.com/segmentio/kafka-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Brokers     []string
	Topic       string
	Producers   int
	MessageSize int
	Compression string
	MetricsPort int
}

type Generator struct {
	config  Config
	writer  *kafka.Writer
	metrics *Metrics
}

type Results struct {
	TotalMessages   int64
	SuccessMessages int64
	FailedMessages  int64
	TotalBytes      int64
	TotalDuration   time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	AvgLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	Latencies       []time.Duration
	Errors          map[string]int64
}

type Metrics struct {
	messagesTotal   *prometheus.CounterVec
	messageDuration *prometheus.HistogramVec
	bytesTotal      prometheus.Counter
	activeProducers prometheus.Gauge
}

func New(config Config) (*Generator, error) {
	var compression kafka.Compression
	switch config.Compression {
	case "gzip":
		compression = kafka.Gzip
	case "snappy":
		compression = kafka.Snappy
	case "lz4":
		compression = kafka.Lz4
	default:
		compression = kafka.None
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Topic:        config.Topic,
		Balancer:     &kafka.LeastBytes{},
		Compression:  compression,
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		MaxAttempts:  3,
	}

	metrics := &Metrics{
		messagesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_load_messages_total",
				Help: "Total number of Kafka messages",
			},
			[]string{"status"},
		),
		messageDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kafka_load_message_duration_seconds",
				Help:    "Kafka message send duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status"},
		),
		bytesTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "kafka_load_bytes_total",
				Help: "Total bytes sent to Kafka",
			},
		),
		activeProducers: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "kafka_load_active_producers",
				Help: "Number of active Kafka producers",
			},
		),
	}

	prometheus.MustRegister(metrics.messagesTotal)
	prometheus.MustRegister(metrics.messageDuration)
	prometheus.MustRegister(metrics.bytesTotal)
	prometheus.MustRegister(metrics.activeProducers)

	return &Generator{
		config:  config,
		writer:  writer,
		metrics: metrics,
	}, nil
}

func (g *Generator) Close() {
	if g.writer != nil {
		g.writer.Close()
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

	msgChan := make(chan struct{}, g.config.Producers*10)

	for i := 0; i < g.config.Producers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.worker(ctx, msgChan, results, &mu)
		}()
	}

	go func() {
		defer close(msgChan)
		currentSecond := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				mps := pattern.RPS(currentSecond)
				for i := 0; i < mps; i++ {
					select {
					case msgChan <- struct{}{}:
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

func (g *Generator) worker(ctx context.Context, msgChan <-chan struct{}, results *Results, mu *sync.Mutex) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-msgChan:
			if !ok {
				return
			}
			g.sendMessage(ctx, results, mu)
		}
	}
}

func (g *Generator) sendMessage(ctx context.Context, results *Results, mu *sync.Mutex) {
	g.metrics.activeProducers.Inc()
	defer g.metrics.activeProducers.Dec()

	atomic.AddInt64(&results.TotalMessages, 1)

	// Generate message
	value := make([]byte, g.config.MessageSize)
	rand.Read(value)

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("key-%d", time.Now().UnixNano())),
		Value: value,
		Time:  time.Now(),
	}

	start := time.Now()
	err := g.writer.WriteMessages(ctx, msg)
	latency := time.Since(start)

	mu.Lock()
	results.Latencies = append(results.Latencies, latency)
	mu.Unlock()

	if err != nil {
		atomic.AddInt64(&results.FailedMessages, 1)
		mu.Lock()
		results.Errors[err.Error()]++
		mu.Unlock()
		g.metrics.messagesTotal.WithLabelValues("error").Inc()
		g.metrics.messageDuration.WithLabelValues("error").Observe(latency.Seconds())
		return
	}

	atomic.AddInt64(&results.SuccessMessages, 1)
	atomic.AddInt64(&results.TotalBytes, int64(g.config.MessageSize))
	g.metrics.messagesTotal.WithLabelValues("success").Inc()
	g.metrics.messageDuration.WithLabelValues("success").Observe(latency.Seconds())
	g.metrics.bytesTotal.Add(float64(g.config.MessageSize))
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
	if r.TotalMessages == 0 {
		return 0
	}
	return float64(r.FailedMessages) / float64(r.TotalMessages)
}

func (r *Results) ToString() string {
	return fmt.Sprintf(`
=== Kafka Load Test Results ===
Total Messages:     %d
Success:            %d (%.2f%%)
Failed:             %d (%.2f%%)
Total Bytes:        %d (%.2f MB)
Duration:           %s
MPS:                %.2f
Throughput:         %.2f MB/s

Latency Statistics:
  Min:              %s
  Max:              %s
  Avg:              %s
  P50:              %s
  P95:              %s
  P99:              %s
`, r.TotalMessages, r.SuccessMessages,
		float64(r.SuccessMessages)/float64(r.TotalMessages)*100,
		r.FailedMessages, r.FailureRate()*100,
		r.TotalBytes, float64(r.TotalBytes)/(1024*1024),
		r.TotalDuration,
		float64(r.TotalMessages)/r.TotalDuration.Seconds(),
		float64(r.TotalBytes)/(1024*1024)/r.TotalDuration.Seconds(),
		r.MinLatency, r.MaxLatency, r.AvgLatency,
		r.P50Latency, r.P95Latency, r.P99Latency)
}

func (r *Results) ToJSON() string {
	return fmt.Sprintf(`{
  "total_messages": %d,
  "success_messages": %d,
  "failed_messages": %d,
  "total_bytes": %d,
  "duration_seconds": %.2f,
  "mps": %.2f,
  "throughput_mbps": %.2f,
  "latency": {
    "min_ms": %.2f,
    "max_ms": %.2f,
    "avg_ms": %.2f,
    "p50_ms": %.2f,
    "p95_ms": %.2f,
    "p99_ms": %.2f
  }
}`, r.TotalMessages, r.SuccessMessages, r.FailedMessages,
		r.TotalBytes,
		r.TotalDuration.Seconds(),
		float64(r.TotalMessages)/r.TotalDuration.Seconds(),
		float64(r.TotalBytes)/(1024*1024)/r.TotalDuration.Seconds(),
		float64(r.MinLatency.Microseconds())/1000,
		float64(r.MaxLatency.Microseconds())/1000,
		float64(r.AvgLatency.Microseconds())/1000,
		float64(r.P50Latency.Microseconds())/1000,
		float64(r.P95Latency.Microseconds())/1000,
		float64(r.P99Latency.Microseconds())/1000)
}
