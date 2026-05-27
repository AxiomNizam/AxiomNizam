package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds conductor/messaging configuration.
type Config struct {
	RabbitMQURL  string
	KafkaBrokers []string
	MaxStream    int
	Backend      string // "rabbitmq", "kafka", or "memory"
	Brokers      string // comma-separated broker addresses
	DefaultTTL   int    // default message TTL in seconds

	// Message buffer settings
	MaxMessages int // max messages kept in memory per topic (default: 10000)

	// Kafka producer settings
	KafkaProducerAcks    int // 0=NoResponse, 1=WaitForLocal, -1=WaitForAll (default: -1)
	KafkaProducerRetries int // max producer retries (default: 3)

	// Stats persistence
	StatsPersistInterval int // persist stats every N messages (default: 10)
}

// DefaultConfig returns production-safe defaults.
func DefaultConfig() Config {
	backend := getEnv("CONDUCTOR_BACKEND", "memory")
	brokers := getEnv("CONDUCTOR_BROKERS", "")

	var rabbitURL string
	var kafkaBrokers []string

	switch strings.ToLower(backend) {
	case "rabbitmq":
		rabbitURL = brokers
	case "kafka":
		if brokers != "" {
			for _, b := range strings.Split(brokers, ",") {
				b = strings.TrimSpace(b)
				if b != "" {
					kafkaBrokers = append(kafkaBrokers, b)
				}
			}
		}
	}

	return Config{
		RabbitMQURL:  rabbitURL,
		KafkaBrokers: kafkaBrokers,
		MaxStream:    500,
		Backend:      backend,
		Brokers:      brokers,
		DefaultTTL:   3600,

		MaxMessages:          10000,
		KafkaProducerAcks:    -1, // WaitForAll
		KafkaProducerRetries: 3,
		StatsPersistInterval: 10,
	}
}

// LoadFromEnv creates a Config from defaults and overrides from env vars.
func LoadFromEnv() Config {
	cfg := DefaultConfig()

	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		cfg.RabbitMQURL = v
	}
	if v := os.Getenv("KAFKA_BROKERS"); v != "" {
		cfg.KafkaBrokers = nil
		for _, b := range strings.Split(v, ",") {
			b = strings.TrimSpace(b)
			if b != "" {
				cfg.KafkaBrokers = append(cfg.KafkaBrokers, b)
			}
		}
	}
	if v := os.Getenv("CONDUCTOR_MAX_STREAM"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			cfg.MaxStream = n
		}
	}
	if v := os.Getenv("CONDUCTOR_DEFAULT_TTL"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			cfg.DefaultTTL = n
		}
	}
	return cfg
}

// Validate checks the configuration for invalid values.
func (c Config) Validate() error {
	switch strings.ToLower(c.Backend) {
	case "rabbitmq", "kafka", "memory":
		// ok
	default:
		return fmt.Errorf("conductor: unknown backend %q (expected rabbitmq, kafka, or memory)", c.Backend)
	}
	if c.Backend != "memory" && len(c.KafkaBrokers) == 0 && c.RabbitMQURL == "" {
		return fmt.Errorf("conductor: brokers or rabbitmq URL required for backend %q", c.Backend)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
