package contracts

// Configurable is the contract for modules that support external configuration.
// Modules implement this to load settings from environment variables, config files,
// or the platform's KV store.
type Configurable interface {
	// LoadFromEnv reads configuration from environment variables.
	// Called once during module initialization before Validate().
	LoadFromEnv() error

	// Validate checks that the loaded configuration is valid.
	// Returns an error describing the first validation failure.
	Validate() error

	// Defaults returns a copy of the config with all fields set to their default values.
	// Used for documentation, testing, and as a fallback when env vars are missing.
	Defaults() Configurable
}

// ConfigProvider is implemented by modules that expose their config for inspection.
// Used by the platform to list active configurations, generate docs, or export to YAML.
type ConfigProvider interface {
	Config() Configurable
}
