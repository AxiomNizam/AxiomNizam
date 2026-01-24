package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents environment configuration
type Config struct {
	values map[string]string
}

// NewConfig creates a new environment config
func NewConfig() *Config {
	return &Config{
		values: make(map[string]string),
	}
}

// Load loads all environment variables
func (c *Config) Load() *Config {
	for _, env := range os.Environ() {
		key, value, _ := strings.Cut(env, "=")
		c.values[key] = value
	}
	return c
}

// Get gets environment variable
func (c *Config) Get(key string) (string, bool) {
	if val, ok := c.values[key]; ok {
		return val, true
	}
	if val, ok := os.LookupEnv(key); ok {
		return val, true
	}
	return "", false
}

// GetString gets environment variable as string
func (c *Config) GetString(key, defaultVal string) string {
	if val, ok := c.Get(key); ok {
		return val
	}
	return defaultVal
}

// GetInt gets environment variable as integer
func (c *Config) GetInt(key string, defaultVal int) int {
	if val, ok := c.Get(key); ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// GetInt64 gets environment variable as int64
func (c *Config) GetInt64(key string, defaultVal int64) int64 {
	if val, ok := c.Get(key); ok {
		if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// GetFloat64 gets environment variable as float64
func (c *Config) GetFloat64(key string, defaultVal float64) float64 {
	if val, ok := c.Get(key); ok {
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			return floatVal
		}
	}
	return defaultVal
}

// GetBool gets environment variable as boolean
func (c *Config) GetBool(key string, defaultVal bool) bool {
	if val, ok := c.Get(key); ok {
		switch strings.ToLower(val) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultVal
}

// GetDuration gets environment variable as duration
func (c *Config) GetDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := c.Get(key); ok {
		if duration, err := time.ParseDuration(val); err == nil {
			return duration
		}
	}
	return defaultVal
}

// GetStringSlice gets environment variable as string slice
func (c *Config) GetStringSlice(key string, sep string) []string {
	if val, ok := c.Get(key); ok {
		return strings.Split(val, sep)
	}
	return []string{}
}

// GetIntSlice gets environment variable as int slice
func (c *Config) GetIntSlice(key string, sep string) []int {
	var ints []int
	if val, ok := c.Get(key); ok {
		for _, str := range strings.Split(val, sep) {
			if intVal, err := strconv.Atoi(str); err == nil {
				ints = append(ints, intVal)
			}
		}
	}
	return ints
}

// Must gets environment variable or panics
func (c *Config) Must(key string) string {
	if val, ok := c.Get(key); ok {
		return val
	}
	panic(fmt.Sprintf("required environment variable not found: %s", key))
}

// Set sets an environment variable in config
func (c *Config) Set(key, value string) *Config {
	c.values[key] = value
	return c
}

// Require requires environment variables
func (c *Config) Require(keys ...string) error {
	for _, key := range keys {
		if _, ok := c.Get(key); !ok {
			return fmt.Errorf("required environment variable not found: %s", key)
		}
	}
	return nil
}

// GetOrError gets environment variable or returns error
func (c *Config) GetOrError(key string) (string, error) {
	if val, ok := c.Get(key); ok {
		return val, nil
	}
	return "", fmt.Errorf("environment variable not found: %s", key)
}

// Environment represents environment type
type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
	Test        Environment = "test"
)

// GetEnvironment gets current environment
func GetEnvironment() Environment {
	envStr := os.Getenv("ENVIRONMENT")
	if envStr == "" {
		envStr = os.Getenv("ENV")
	}
	if envStr == "" {
		envStr = "development"
	}

	switch strings.ToLower(envStr) {
	case "prod", "production":
		return Production
	case "staging":
		return Staging
	case "test":
		return Test
	default:
		return Development
	}
}

// IsDevelopment checks if environment is development
func IsDevelopment() bool {
	return GetEnvironment() == Development
}

// IsStaging checks if environment is staging
func IsStaging() bool {
	return GetEnvironment() == Staging
}

// IsProduction checks if environment is production
func IsProduction() bool {
	return GetEnvironment() == Production
}

// IsTest checks if environment is test
func IsTest() bool {
	return GetEnvironment() == Test
}

// IsLocal checks if running locally
func IsLocal() bool {
	return !IsProduction()
}

// EnvConfig loads environment-specific configuration
type EnvConfig struct {
	Environment Environment
	config      *Config
}

// NewEnvConfig creates new environment config loader
func NewEnvConfig() *EnvConfig {
	return &EnvConfig{
		Environment: GetEnvironment(),
		config:      NewConfig().Load(),
	}
}

// GetString gets string with environment awareness
func (ec *EnvConfig) GetString(key, defaultVal string) string {
	// Try environment-specific key first
	envKey := fmt.Sprintf("%s_%s", strings.ToUpper(string(ec.Environment)), key)
	if val, ok := ec.config.Get(envKey); ok {
		return val
	}
	return ec.config.GetString(key, defaultVal)
}

// GetInt gets int with environment awareness
func (ec *EnvConfig) GetInt(key string, defaultVal int) int {
	// Try environment-specific key first
	envKey := fmt.Sprintf("%s_%s", strings.ToUpper(string(ec.Environment)), key)
	if val, ok := ec.config.Get(envKey); ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return ec.config.GetInt(key, defaultVal)
}

// GetBool gets bool with environment awareness
func (ec *EnvConfig) GetBool(key string, defaultVal bool) bool {
	// Try environment-specific key first
	envKey := fmt.Sprintf("%s_%s", strings.ToUpper(string(ec.Environment)), key)
	if val, ok := ec.config.Get(envKey); ok {
		switch strings.ToLower(val) {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		}
	}
	return ec.config.GetBool(key, defaultVal)
}

// Parser represents environment variable parser
type Parser struct {
	env *Config
}

// NewParser creates a new parser
func NewParser() *Parser {
	return &Parser{
		env: NewConfig().Load(),
	}
}

// Parse parses environment variables into struct
func (p *Parser) Parse(prefix string) map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		key, value, _ := strings.Cut(env, "=")
		if strings.HasPrefix(key, prefix) {
			result[key] = value
		}
	}
	return result
}

// Validator validates required environment variables
type Validator struct {
	required []string
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		required: make([]string, 0),
	}
}

// Require marks a variable as required
func (v *Validator) Require(key string) *Validator {
	v.required = append(v.required, key)
	return v
}

// Validate validates environment variables
func (v *Validator) Validate() error {
	for _, key := range v.required {
		if _, ok := os.LookupEnv(key); !ok {
			return fmt.Errorf("required environment variable not set: %s", key)
		}
	}
	return nil
}

// GetServiceConfig gets service configuration from environment
func GetServiceConfig() ServiceConfig {
	config := NewConfig().Load()
	return ServiceConfig{
		Name:    config.GetString("SERVICE_NAME", "axiom-nizam"),
		Version: config.GetString("SERVICE_VERSION", "0.1.0"),
		Port:    config.GetInt("SERVICE_PORT", 8080),
		Debug:   config.GetBool("DEBUG", false),
	}
}

// ServiceConfig represents service configuration
type ServiceConfig struct {
	Name    string
	Version string
	Port    int
	Debug   bool
}
