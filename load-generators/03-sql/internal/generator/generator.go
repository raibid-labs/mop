package generator

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/raibid-labs/mop/load-generators/03-sql/internal/patterns"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	DBHost      string
	DBPort      int
	DBUser      string
	DBPassword  string
	DBName      string
	Clients     int
	QueryType   string
	MetricsPort int
}

type Generator struct {
	config  Config
	db      *sql.DB
	metrics *Metrics
}

type Results struct {
	TotalTransactions int64
	SuccessTransactions int64
	FailedTransactions  int64
	TotalDuration      time.Duration
	MinLatency         time.Duration
	MaxLatency         time.Duration
	AvgLatency         time.Duration
	P50Latency         time.Duration
	P95Latency         time.Duration
	P99Latency         time.Duration
	Latencies          []time.Duration
	Errors             map[string]int64
}

type Metrics struct {
	transactionsTotal *prometheus.CounterVec
	queryDuration     *prometheus.HistogramVec
	activeConnections prometheus.Gauge
}

func New(config Config) (*Generator, error) {
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(config.Clients * 2)
	db.SetMaxIdleConns(config.Clients)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	metrics := &Metrics{
		transactionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sql_load_transactions_total",
				Help: "Total number of SQL transactions",
			},
			[]string{"status", "type"},
		),
		queryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "sql_load_query_duration_seconds",
				Help:    "SQL query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status", "type"},
		),
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "sql_load_active_connections",
				Help: "Number of active database connections",
			},
		),
	}

	prometheus.MustRegister(metrics.transactionsTotal)
	prometheus.MustRegister(metrics.queryDuration)
	prometheus.MustRegister(metrics.activeConnections)

	return &Generator{
		config:  config,
		db:      db,
		metrics: metrics,
	}, nil
}

func (g *Generator) Close() {
	if g.db != nil {
		g.db.Close()
	}
}

func (g *Generator) Initialize(ctx context.Context) error {
	// Create load test table if not exists
	_, err := g.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS load_test_orders (
			id SERIAL PRIMARY KEY,
			customer_id INT NOT NULL,
			product_name VARCHAR(255) NOT NULL,
			quantity INT NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create index
	_, err = g.db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_customer_id ON load_test_orders(customer_id)
	`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
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

	txChan := make(chan struct{}, g.config.Clients*10)

	// Start workers
	for i := 0; i < g.config.Clients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.worker(ctx, txChan, results, &mu)
		}()
	}

	// Generate load according to pattern
	go func() {
		defer close(txChan)
		currentSecond := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				tps := pattern.RPS(currentSecond)
				for i := 0; i < tps; i++ {
					select {
					case txChan <- struct{}{}:
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

func (g *Generator) worker(ctx context.Context, txChan <-chan struct{}, results *Results, mu *sync.Mutex) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-txChan:
			if !ok {
				return
			}
			g.executeTransaction(ctx, results, mu)
		}
	}
}

func (g *Generator) executeTransaction(ctx context.Context, results *Results, mu *sync.Mutex) {
	g.metrics.activeConnections.Inc()
	defer g.metrics.activeConnections.Dec()

	atomic.AddInt64(&results.TotalTransactions, 1)

	// Determine query type
	queryType := g.config.QueryType
	if queryType == "mixed" {
		if rand.Float64() < 0.7 {
			queryType = "read"
		} else {
			queryType = "write"
		}
	}

	start := time.Now()
	var err error

	switch queryType {
	case "read":
		err = g.executeRead(ctx)
	case "write":
		err = g.executeWrite(ctx)
	}

	latency := time.Since(start)

	mu.Lock()
	results.Latencies = append(results.Latencies, latency)
	mu.Unlock()

	if err != nil {
		atomic.AddInt64(&results.FailedTransactions, 1)
		mu.Lock()
		results.Errors[err.Error()]++
		mu.Unlock()
		g.metrics.transactionsTotal.WithLabelValues("error", queryType).Inc()
		g.metrics.queryDuration.WithLabelValues("error", queryType).Observe(latency.Seconds())
		return
	}

	atomic.AddInt64(&results.SuccessTransactions, 1)
	g.metrics.transactionsTotal.WithLabelValues("success", queryType).Inc()
	g.metrics.queryDuration.WithLabelValues("success", queryType).Observe(latency.Seconds())
}

func (g *Generator) executeRead(ctx context.Context) error {
	// Random read queries
	queryType := rand.Intn(3)

	switch queryType {
	case 0:
		// SELECT by customer_id
		var count int
		customerID := rand.Intn(1000) + 1
		err := g.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM load_test_orders WHERE customer_id = $1",
			customerID).Scan(&count)
		return err

	case 1:
		// SELECT recent orders
		rows, err := g.db.QueryContext(ctx,
			"SELECT id, customer_id, product_name, quantity, price FROM load_test_orders ORDER BY created_at DESC LIMIT 10")
		if err != nil {
			return err
		}
		defer rows.Close()
		return nil

	case 2:
		// Aggregate query
		var total float64
		err := g.db.QueryRowContext(ctx,
			"SELECT COALESCE(SUM(price * quantity), 0) FROM load_test_orders WHERE status = 'completed'").Scan(&total)
		return err
	}

	return nil
}

func (g *Generator) executeWrite(ctx context.Context) error {
	// Random write operations
	writeType := rand.Intn(2)

	switch writeType {
	case 0:
		// INSERT
		customerID := rand.Intn(1000) + 1
		products := []string{"Widget", "Gadget", "Doohickey", "Thingamajig", "Whatsit"}
		productName := products[rand.Intn(len(products))]
		quantity := rand.Intn(10) + 1
		price := rand.Float64() * 100

		_, err := g.db.ExecContext(ctx,
			"INSERT INTO load_test_orders (customer_id, product_name, quantity, price, status) VALUES ($1, $2, $3, $4, $5)",
			customerID, productName, quantity, price, "pending")
		return err

	case 1:
		// UPDATE
		orderID := rand.Intn(1000) + 1
		statuses := []string{"pending", "processing", "completed", "cancelled"}
		status := statuses[rand.Intn(len(statuses))]

		_, err := g.db.ExecContext(ctx,
			"UPDATE load_test_orders SET status = $1 WHERE id = $2",
			status, orderID)
		return err
	}

	return nil
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
	if r.TotalTransactions == 0 {
		return 0
	}
	return float64(r.FailedTransactions) / float64(r.TotalTransactions)
}

func (r *Results) ToString() string {
	return fmt.Sprintf(`
=== SQL Load Test Results ===
Total Transactions: %d
Success:            %d (%.2f%%)
Failed:             %d (%.2f%%)
Duration:           %s
TPS:                %.2f

Latency Statistics:
  Min:              %s
  Max:              %s
  Avg:              %s
  P50:              %s
  P95:              %s
  P99:              %s
`, r.TotalTransactions, r.SuccessTransactions,
		float64(r.SuccessTransactions)/float64(r.TotalTransactions)*100,
		r.FailedTransactions, r.FailureRate()*100,
		r.TotalDuration,
		float64(r.TotalTransactions)/r.TotalDuration.Seconds(),
		r.MinLatency, r.MaxLatency, r.AvgLatency,
		r.P50Latency, r.P95Latency, r.P99Latency)
}

func (r *Results) ToJSON() string {
	return fmt.Sprintf(`{
  "total_transactions": %d,
  "success_transactions": %d,
  "failed_transactions": %d,
  "duration_seconds": %.2f,
  "tps": %.2f,
  "latency": {
    "min_ms": %.2f,
    "max_ms": %.2f,
    "avg_ms": %.2f,
    "p50_ms": %.2f,
    "p95_ms": %.2f,
    "p99_ms": %.2f
  }
}`, r.TotalTransactions, r.SuccessTransactions, r.FailedTransactions,
		r.TotalDuration.Seconds(),
		float64(r.TotalTransactions)/r.TotalDuration.Seconds(),
		float64(r.MinLatency.Microseconds())/1000,
		float64(r.MaxLatency.Microseconds())/1000,
		float64(r.AvgLatency.Microseconds())/1000,
		float64(r.P50Latency.Microseconds())/1000,
		float64(r.P95Latency.Microseconds())/1000,
		float64(r.P99Latency.Microseconds())/1000)
}
