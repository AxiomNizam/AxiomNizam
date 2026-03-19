package waitx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// MultiChecker combines checkers and verifies all are ready.
type MultiChecker struct {
	NamePrefix string
	Checkers   []Checker
	Parallel   bool
}

func (c MultiChecker) Name() string {
	if strings.TrimSpace(c.NamePrefix) != "" {
		return strings.TrimSpace(c.NamePrefix)
	}
	return "multi-checker"
}

func (c MultiChecker) Check(ctx context.Context) error {
	if len(c.Checkers) == 0 {
		return fmt.Errorf("at least one checker is required")
	}

	if c.Parallel {
		return c.checkParallel(ctx)
	}

	for _, checker := range c.Checkers {
		if checker == nil {
			continue
		}
		if err := checker.Check(ctx); err != nil {
			return fmt.Errorf("%s failed: %w", checker.Name(), err)
		}
	}

	return nil
}

func (c MultiChecker) checkParallel(ctx context.Context) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errs := make([]error, 0)

	for _, checker := range c.Checkers {
		if checker == nil {
			continue
		}
		wg.Add(1)
		go func(chk Checker) {
			defer wg.Done()
			if err := chk.Check(ctx); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s failed: %w", chk.Name(), err))
				mu.Unlock()
			}
		}(checker)
	}

	wg.Wait()
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// RedisChecker validates Redis connectivity and optional key/value checks.
type RedisChecker struct {
	Address            string
	ExpectedKey        string
	ExpectedValueRegex *regexp.Regexp
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
}

func (c RedisChecker) Name() string {
	return "redis:" + strings.TrimSpace(c.Address)
}

func (c RedisChecker) Check(ctx context.Context) error {
	address := strings.TrimSpace(c.Address)
	if address == "" {
		return fmt.Errorf("redis address is required")
	}

	options, err := redisOptionsFromAddress(address)
	if err != nil {
		return err
	}

	if c.DialTimeout > 0 {
		options.DialTimeout = c.DialTimeout
	}
	if c.ReadTimeout > 0 {
		options.ReadTimeout = c.ReadTimeout
	}

	client := redis.NewClient(options)
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	key := strings.TrimSpace(c.ExpectedKey)
	if key == "" {
		return nil
	}

	value, err := client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("redis key %q not found", key)
	}
	if err != nil {
		return fmt.Errorf("redis get key %q failed: %w", key, err)
	}

	if c.ExpectedValueRegex != nil && !c.ExpectedValueRegex.MatchString(value) {
		return fmt.Errorf("redis key %q value did not match regex %q", key, c.ExpectedValueRegex.String())
	}

	return nil
}

// MySQLChecker validates MySQL connectivity and optional table existence.
type MySQLChecker struct {
	DSN            string
	ExpectedTable  string
	ConnectTimeout time.Duration
}

func (c MySQLChecker) Name() string {
	return "mysql"
}

func (c MySQLChecker) Check(ctx context.Context) error {
	dsn := strings.TrimSpace(c.DSN)
	if dsn == "" {
		return fmt.Errorf("mysql dsn is required")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("mysql open failed: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	queryCtx, cancel := withOptionalTimeout(ctx, c.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(queryCtx); err != nil {
		return fmt.Errorf("mysql ping failed: %w", err)
	}

	table := strings.TrimSpace(c.ExpectedTable)
	if table == "" {
		return nil
	}

	const tableQuery = "SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ? LIMIT 1"
	var exists int
	err = db.QueryRowContext(queryCtx, tableQuery, table).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("mysql table %q not found", table)
	}
	if err != nil {
		return fmt.Errorf("mysql table lookup failed: %w", err)
	}

	return nil
}

// PostgreSQLChecker validates PostgreSQL connectivity and optional table existence.
type PostgreSQLChecker struct {
	DSN            string
	ExpectedTable  string
	ConnectTimeout time.Duration
}

func (c PostgreSQLChecker) Name() string {
	return "postgresql"
}

func (c PostgreSQLChecker) Check(ctx context.Context) error {
	dsn := strings.TrimSpace(c.DSN)
	if dsn == "" {
		return fmt.Errorf("postgresql dsn is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("postgresql open failed: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	queryCtx, cancel := withOptionalTimeout(ctx, c.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(queryCtx); err != nil {
		return fmt.Errorf("postgresql ping failed: %w", err)
	}

	table := strings.TrimSpace(c.ExpectedTable)
	if table == "" {
		return nil
	}

	const tableQuery = "SELECT 1 FROM information_schema.tables WHERE table_name = $1 AND table_schema = ANY(current_schemas(false)) LIMIT 1"
	var exists int
	err = db.QueryRowContext(queryCtx, tableQuery, table).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("postgresql table %q not found in current schemas", table)
	}
	if err != nil {
		return fmt.Errorf("postgresql table lookup failed: %w", err)
	}

	return nil
}

// MongoDBChecker validates MongoDB connectivity.
type MongoDBChecker struct {
	URI            string
	ConnectTimeout time.Duration
}

func (c MongoDBChecker) Name() string {
	return "mongodb"
}

func (c MongoDBChecker) Check(ctx context.Context) error {
	uri := strings.TrimSpace(c.URI)
	if uri == "" {
		return fmt.Errorf("mongodb uri is required")
	}

	connectCtx, cancel := withOptionalTimeout(ctx, c.ConnectTimeout)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	if c.ConnectTimeout > 0 {
		clientOpts.SetConnectTimeout(c.ConnectTimeout)
	}

	client, err := mongo.Connect(connectCtx, clientOpts)
	if err != nil {
		return fmt.Errorf("mongodb connect failed: %w", err)
	}
	defer client.Disconnect(context.Background())

	if err := client.Ping(connectCtx, readpref.Primary()); err != nil {
		return fmt.Errorf("mongodb ping failed: %w", err)
	}

	return nil
}

// RabbitMQChecker validates RabbitMQ endpoint connectivity.
type RabbitMQChecker struct {
	URL         string
	DialTimeout time.Duration
}

func (c RabbitMQChecker) Name() string {
	return "rabbitmq:" + strings.TrimSpace(c.URL)
}

func (c RabbitMQChecker) Check(_ context.Context) error {
	address, err := addressFromTarget(strings.TrimSpace(c.URL), "5672")
	if err != nil {
		return fmt.Errorf("rabbitmq target parse failed: %w", err)
	}
	if err := dialAddress(address, c.DialTimeout); err != nil {
		return fmt.Errorf("rabbitmq dial failed: %w", err)
	}
	return nil
}

// KafkaChecker validates one or more Kafka broker endpoints.
type KafkaChecker struct {
	Brokers     []string
	DialTimeout time.Duration
}

func (c KafkaChecker) Name() string {
	return "kafka"
}

func (c KafkaChecker) Check(_ context.Context) error {
	brokers := sanitizeStringSlice(c.Brokers)
	if len(brokers) == 0 {
		return fmt.Errorf("at least one kafka broker is required")
	}

	for _, broker := range brokers {
		address, err := addressFromTarget(broker, "9092")
		if err != nil {
			return fmt.Errorf("kafka broker parse failed for %q: %w", broker, err)
		}
		if err := dialAddress(address, c.DialTimeout); err != nil {
			return fmt.Errorf("kafka broker %q dial failed: %w", address, err)
		}
	}

	return nil
}

// InfluxDBChecker validates InfluxDB HTTP endpoint readiness.
type InfluxDBChecker struct {
	URL              string
	RequestTimeout   time.Duration
	ExpectStatusCode int
}

func (c InfluxDBChecker) Name() string {
	return "influxdb:" + strings.TrimSpace(c.URL)
}

func (c InfluxDBChecker) Check(ctx context.Context) error {
	endpoint := strings.TrimSpace(c.URL)
	if endpoint == "" {
		return fmt.Errorf("influxdb url is required")
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("influxdb url parse failed: %w", err)
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "http"
	}
	if parsed.Host == "" {
		return fmt.Errorf("influxdb url must include host")
	}
	if parsed.Path == "" || parsed.Path == "/" {
		parsed.Path = "/health"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return fmt.Errorf("influxdb request build failed: %w", err)
	}

	client := &http.Client{}
	if c.RequestTimeout > 0 {
		client.Timeout = c.RequestTimeout
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("influxdb request failed: %w", err)
	}
	defer resp.Body.Close()

	expected := c.ExpectStatusCode
	if expected > 0 {
		if resp.StatusCode != expected {
			return fmt.Errorf("unexpected influxdb status code: got %d expected %d", resp.StatusCode, expected)
		}
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected influxdb status code: %d", resp.StatusCode)
	}

	return nil
}

// TemporalServerChecker validates Temporal server endpoint reachability.
type TemporalServerChecker struct {
	Target      string
	DialTimeout time.Duration
}

func (c TemporalServerChecker) Name() string {
	return "temporal:" + strings.TrimSpace(c.Target)
}

func (c TemporalServerChecker) Check(_ context.Context) error {
	address, err := addressFromTarget(strings.TrimSpace(c.Target), "7233")
	if err != nil {
		return fmt.Errorf("temporal target parse failed: %w", err)
	}
	if err := dialAddress(address, c.DialTimeout); err != nil {
		return fmt.Errorf("temporal dial failed: %w", err)
	}
	return nil
}

func redisOptionsFromAddress(address string) (*redis.Options, error) {
	if strings.Contains(address, "://") {
		options, err := redis.ParseURL(address)
		if err != nil {
			return nil, fmt.Errorf("redis url parse failed: %w", err)
		}
		return options, nil
	}

	addr := ensurePort(address, "6379")
	if strings.TrimSpace(addr) == "" {
		return nil, fmt.Errorf("redis address is required")
	}

	return &redis.Options{Addr: addr}, nil
}

func addressFromTarget(target, defaultPort string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("target is required")
	}

	if strings.Contains(target, "://") {
		parsed, err := url.Parse(target)
		if err != nil {
			return "", err
		}
		if parsed.Host == "" {
			return "", fmt.Errorf("target must include host")
		}
		return ensurePort(parsed.Host, defaultPort), nil
	}

	return ensurePort(target, defaultPort), nil
}

func dialAddress(address string, dialTimeout time.Duration) error {
	timeout := dialTimeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
