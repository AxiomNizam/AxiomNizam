package client

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the AxiomNizam CLI configuration
type Config struct {
	CurrentContext string       `yaml:"current-context"`
	Contexts       []Context    `yaml:"contexts"`
	APIVersion     string       `yaml:"apiVersion,omitempty"`
	Kind           string       `yaml:"kind,omitempty"`
}

// Context represents a context configuration
type Context struct {
	Name    string           `yaml:"name"`
	Cluster *ClusterInfo     `yaml:"cluster"`
	User    string           `yaml:"user"`
	Namespace string         `yaml:"namespace,omitempty"`
}

// ClusterInfo contains cluster details
type ClusterInfo struct {
	Server                   string `yaml:"server"`
	InsecureSkipTLSVerify    bool   `yaml:"insecure-skip-tls-verify,omitempty"`
	CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
}

// ConfigManager manages CLI configuration
type ConfigManager struct {
	configPath string
	config     *Config
}

// DefaultConfigPath returns the default config path
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".axiomnizam", "config.yaml")
}

// NewConfigManager creates a new config manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		configPath: DefaultConfigPath(),
	}
}

// Load loads the configuration from file
func (cm *ConfigManager) Load() error {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Create default config if not exists
		cm.config = &Config{
			APIVersion: "axiom-nizam/v1",
			Kind:       "Config",
			CurrentContext: "default",
			Contexts: []Context{
				{
					Name: "default",
					Cluster: &ClusterInfo{
						Server: "http://localhost:8000",
					},
					User: "default",
					Namespace: "default",
				},
			},
		}
		return cm.Save()
	}

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	cm.config = &Config{}
	if err := yaml.Unmarshal(data, cm.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// Save saves the configuration to file
func (cm *ConfigManager) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cm.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetCurrentContext returns the current context
func (cm *ConfigManager) GetCurrentContext() *Context {
	if cm.config == nil {
		return nil
	}

	for i := range cm.config.Contexts {
		if cm.config.Contexts[i].Name == cm.config.CurrentContext {
			return &cm.config.Contexts[i]
		}
	}

	if len(cm.config.Contexts) > 0 {
		return &cm.config.Contexts[0]
	}

	return nil
}

// SetCurrentContext sets the current context
func (cm *ConfigManager) SetCurrentContext(contextName string) error {
	if cm.config == nil {
		return fmt.Errorf("config not loaded")
	}

	// Verify context exists
	for _, ctx := range cm.config.Contexts {
		if ctx.Name == contextName {
			cm.config.CurrentContext = contextName
			return cm.Save()
		}
	}

	return fmt.Errorf("context not found: %s", contextName)
}

// GetCluster returns cluster info for the current context
func (cm *ConfigManager) GetCluster() *ClusterInfo {
	ctx := cm.GetCurrentContext()
	if ctx == nil {
		return nil
	}
	return ctx.Cluster
}

// GetServer returns the server URL
func (cm *ConfigManager) GetServer() string {
	cluster := cm.GetCluster()
	if cluster == nil {
		return "http://localhost:8000"
	}
	return cluster.Server
}

// GetNamespace returns the namespace for current context
func (cm *ConfigManager) GetNamespace() string {
	ctx := cm.GetCurrentContext()
	if ctx == nil {
		return "default"
	}
	if ctx.Namespace == "" {
		return "default"
	}
	return ctx.Namespace
}

// GetToken returns the auth token from file
func (cm *ConfigManager) GetToken() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	tokenPath := filepath.Join(home, ".axiomnizam", "token")
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", nil // Token file doesn't exist yet
	}

	return string(data), nil
}

// SetToken saves the auth token to file
func (cm *ConfigManager) SetToken(token string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	tokenDir := filepath.Join(home, ".axiomnizam")
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	tokenPath := filepath.Join(tokenDir, "token")
	if err := os.WriteFile(tokenPath, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// DeleteToken deletes the stored token
func (cm *ConfigManager) DeleteToken() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	tokenPath := filepath.Join(home, ".axiomnizam", "token")
	if _, err := os.Stat(tokenPath); err == nil {
		return os.Remove(tokenPath)
	}

	return nil
}

// AddContext adds a new context
func (cm *ConfigManager) AddContext(context Context) error {
	if cm.config == nil {
		return fmt.Errorf("config not loaded")
	}

	// Check if context already exists
	for i, ctx := range cm.config.Contexts {
		if ctx.Name == context.Name {
			cm.config.Contexts[i] = context
			return cm.Save()
		}
	}

	cm.config.Contexts = append(cm.config.Contexts, context)
	return cm.Save()
}

// ListContexts returns all available contexts
func (cm *ConfigManager) ListContexts() []Context {
	if cm.config == nil {
		return []Context{}
	}
	return cm.config.Contexts
}

// GetConfigPath returns the config file path
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}
