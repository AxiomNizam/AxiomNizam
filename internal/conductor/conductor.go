package conductor

import "example.com/axiomnizam/internal/conductor/config"

// LoadConfigFromEnv re-exports the config loader from the config subpackage.
var LoadConfigFromEnv = config.LoadFromEnv

// DefaultConfig re-exports the default config constructor.
var DefaultConfig = config.DefaultConfig
