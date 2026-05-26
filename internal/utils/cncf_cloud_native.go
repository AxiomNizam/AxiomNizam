package utils

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// AwsConfig holds AWS configuration
type AwsConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Endpoint        string
}

// AzureConfig holds Azure configuration
type AzureConfig struct {
	SubscriptionID string
	TenantID       string
	ClientID       string
	ClientSecret   string
	ResourceGroup  string
}

// GcpConfig holds GCP configuration
type GcpConfig struct {
	ProjectID string
	CredPath  string
	Zone      string
	Region    string
}

// CloudProvider defines the interface for cloud provider operations
type CloudProvider interface {
	GetRegion() string
	GetEndpoint() string
	Authenticate() error
	ListResources(resourceType string) ([]string, error)
	CreateResource(resourceType, resourceName string) error
	DeleteResource(resourceType, resourceName string) error
}

// ObservabilityStack holds observability components configuration
type ObservabilityStack struct {
	Prometheus   *PrometheusConfig
	Grafana      *GrafanaConfig
	Loki         *LokiConfig
	Jaeger       *JaegerConfig
	AlertManager *AlertManagerConfig
}

// PrometheusConfig holds Prometheus configuration
type PrometheusConfig struct {
	Endpoint           string
	ScrapeInterval     time.Duration
	EvaluationInterval time.Duration
	GlobalLabels       map[string]string
	AlertRules         []string
}

// NewPrometheusConfig creates a new Prometheus configuration.
// Endpoint defaults to env GRAFANA_PROMETHEUS_URL or "http://prometheus:9090".
func NewPrometheusConfig() *PrometheusConfig {
	endpoint := os.Getenv("PROMETHEUS_URL")
	if endpoint == "" {
		endpoint = "http://prometheus:9090"
	}
	return &PrometheusConfig{
		Endpoint:           endpoint,
		ScrapeInterval:     15 * time.Second,
		EvaluationInterval: 15 * time.Second,
		GlobalLabels:       make(map[string]string),
		AlertRules:         make([]string, 0),
	}
}

// GrafanaConfig holds Grafana configuration
type GrafanaConfig struct {
	Endpoint         string
	AdminUsername    string
	AdminPassword    string
	DefaultDashboard string
	DataSource       string
}

// NewGrafanaConfig creates a new Grafana configuration.
// Credentials default to env GRAFANA_ADMIN_USER / GRAFANA_ADMIN_PASSWORD.
func NewGrafanaConfig() *GrafanaConfig {
	user := os.Getenv("GRAFANA_ADMIN_USER")
	if user == "" {
		user = "admin"
	}
	pass := os.Getenv("GRAFANA_ADMIN_PASSWORD")
	if pass == "" {
		pass = "admin"
	}
	endpoint := os.Getenv("GRAFANA_URL")
	if endpoint == "" {
		endpoint = "http://grafana:3000"
	}
	return &GrafanaConfig{
		Endpoint:         endpoint,
		AdminUsername:    user,
		AdminPassword:    pass,
		DefaultDashboard: "axiom-nizam",
		DataSource:       "Prometheus",
	}
}

// LokiConfig holds Loki configuration
type LokiConfig struct {
	Endpoint          string
	RetentionDays     int
	MaxChunkAge       time.Duration
	ChunkScanInterval time.Duration
}

// NewLokiConfig creates a new Loki configuration.
// Endpoint defaults to env LOKI_URL or "http://loki:3100".
func NewLokiConfig() *LokiConfig {
	endpoint := os.Getenv("LOKI_URL")
	if endpoint == "" {
		endpoint = "http://loki:3100"
	}
	return &LokiConfig{
		Endpoint:          endpoint,
		RetentionDays:     30,
		MaxChunkAge:       2 * time.Hour,
		ChunkScanInterval: 10 * time.Minute,
	}
}

// JaegerConfig holds Jaeger configuration
type JaegerConfig struct {
	Endpoint     string
	AgentHost    string
	AgentPort    int
	SamplingRate float64
	BufferSize   int
}

// NewJaegerConfig creates a new Jaeger configuration.
// Endpoint defaults to env JAEGER_URL or "http://jaeger:6831".
func NewJaegerConfig() *JaegerConfig {
	endpoint := os.Getenv("JAEGER_URL")
	if endpoint == "" {
		endpoint = "http://jaeger:6831"
	}
	return &JaegerConfig{
		Endpoint:     endpoint,
		AgentHost:    "jaeger",
		AgentPort:    6831,
		SamplingRate: 0.1,
		BufferSize:   1000,
	}
}

// AlertManagerConfig holds AlertManager configuration
type AlertManagerConfig struct {
	Endpoint    string
	Receivers   []AlertReceiver
	Routes      []AlertRoute
	Inhibitions []AlertInhibition
}

// AlertReceiver defines an alert receiver
type AlertReceiver struct {
	Name   string
	Type   string // email, slack, pagerduty, etc
	Config map[string]interface{}
}

// AlertRoute defines an alert route
type AlertRoute struct {
	Matcher  string
	Receiver string
	GroupBy  []string
	Continue bool
}

// AlertInhibition defines alert inhibition
type AlertInhibition struct {
	SourceMatch map[string]string
	TargetMatch map[string]string
	Equal       []string
}

// NewAlertManagerConfig creates a new AlertManager configuration
func NewAlertManagerConfig() *AlertManagerConfig {
	return &AlertManagerConfig{
		Endpoint:    "http://alertmanager:9093",
		Receivers:   make([]AlertReceiver, 0),
		Routes:      make([]AlertRoute, 0),
		Inhibitions: make([]AlertInhibition, 0),
	}
}

// ServiceMeshConfig holds service mesh configuration
type ServiceMeshConfig struct {
	Type    string // istio, linkerd, etc
	Version string
	Config  map[string]interface{}
}

// ContainerRegistry holds container registry configuration
type ContainerRegistry struct {
	Registry string
	Username string
	Password string
	Insecure bool
}

// ImageBuilder handles container image building
type ImageBuilder struct {
	Dockerfile string
	BuildArgs  map[string]string
	Labels     map[string]string
	Registry   *ContainerRegistry
}

// NewImageBuilder creates a new image builder
func NewImageBuilder(dockerfile string) *ImageBuilder {
	return &ImageBuilder{
		Dockerfile: dockerfile,
		BuildArgs:  make(map[string]string),
		Labels:     make(map[string]string),
	}
}

// BuildImage builds a container image
func (ib *ImageBuilder) BuildImage(imageName, tag string) (string, error) {
	if imageName == "" {
		return "", fmt.Errorf("image name cannot be empty")
	}
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("%s:%s", imageName, tag), nil
}

// PushImage pushes an image to registry
func (ib *ImageBuilder) PushImage(imageName string) error {
	if ib.Registry == nil {
		return fmt.Errorf("no registry configured")
	}
	fmt.Printf("Pushing %s to %s\n", imageName, ib.Registry.Registry)
	return nil
}

// GitOpsConfig holds GitOps configuration
type GitOpsConfig struct {
	Repository    string
	Branch        string
	Path          string
	SyncInterval  time.Duration
	RetryInterval time.Duration
	Prune         bool
	SelfHeal      bool
}

// NewGitOpsConfig creates a new GitOps configuration
func NewGitOpsConfig() *GitOpsConfig {
	return &GitOpsConfig{
		Branch:        "main",
		SyncInterval:  3 * time.Minute,
		RetryInterval: 30 * time.Second,
		Prune:         true,
		SelfHeal:      true,
	}
}

// IngressController holds ingress controller configuration
type IngressController struct {
	ClassName    string
	IngressClass string
	Controller   string
	Config       map[string]interface{}
}

// Ingress defines an ingress configuration
type Ingress struct {
	Name      string
	Namespace string
	Rules     []IngressRule
	TLS       []IngressTLS
}

// IngressRule defines an ingress rule
type IngressRule struct {
	Host  string
	Paths []IngressPath
}

// IngressPath defines an ingress path
type IngressPath struct {
	Path        string
	PathType    string // Prefix, Exact, ImplementationSpecific
	ServiceName string
	ServicePort int
}

// IngressTLS defines TLS configuration for ingress
type IngressTLS struct {
	Hosts      []string
	SecretName string
}

// NewIngress creates a new ingress
func NewIngress(name, namespace string) *Ingress {
	return &Ingress{
		Name:      name,
		Namespace: namespace,
		Rules:     make([]IngressRule, 0),
		TLS:       make([]IngressTLS, 0),
	}
}

// HelmRelease holds Helm release configuration
type HelmRelease struct {
	Name      string
	Chart     string
	Version   string
	Namespace string
	Values    map[string]interface{}
	Hooks     []HelmHook
}

// HelmHook defines a Helm hook
type HelmHook struct {
	Type       string // pre-install, post-install, pre-upgrade, post-upgrade, pre-delete, post-delete
	Annotation string
	Weight     int
}

// NewHelmRelease creates a new Helm release
func NewHelmRelease(name, chart, namespace string) *HelmRelease {
	return &HelmRelease{
		Name:      name,
		Chart:     chart,
		Namespace: namespace,
		Values:    make(map[string]interface{}),
		Hooks:     make([]HelmHook, 0),
	}
}

// OperatorConfig holds Kubernetes operator configuration
type OperatorConfig struct {
	Name      string
	Version   string
	Namespace string
	RBAC      bool
	Webhooks  []WebhookConfig
}

// WebhookConfig defines a webhook configuration
type WebhookConfig struct {
	Type          string // ValidatingWebhook, MutatingWebhook
	FailurePolicy string // Fail, Ignore
	SideEffect    string // None, Some, NoneOnDryRun
	Rules         []WebhookRule
}

// WebhookRule defines a webhook rule
type WebhookRule struct {
	Operations  []string
	APIGroups   []string
	APIVersions []string
	Resources   []string
	Scope       string // "*", "Namespaced", "Cluster"
}

// NewOperatorConfig creates a new operator configuration
func NewOperatorConfig(name, namespace string) *OperatorConfig {
	return &OperatorConfig{
		Name:      name,
		Namespace: namespace,
		RBAC:      true,
		Webhooks:  make([]WebhookConfig, 0),
	}
}

// FluxConfig holds Flux CD configuration
type FluxConfig struct {
	Namespace     string
	Version       string
	GitRepository GitRepositoryConfig
	Kustomization KustomizationConfig
}

// GitRepositoryConfig holds Git repository configuration
type GitRepositoryConfig struct {
	URL      string
	Interval time.Duration
	Ref      string
}

// KustomizationConfig holds Kustomization configuration
type KustomizationConfig struct {
	Path     string
	Interval time.Duration
	Prune    bool
	Health   bool
}

// NewFluxConfig creates a new Flux configuration
func NewFluxConfig() *FluxConfig {
	return &FluxConfig{
		Namespace: "flux-system",
		Version:   "2.0.0",
		GitRepository: GitRepositoryConfig{
			Interval: 1 * time.Minute,
			Ref:      "main",
		},
		Kustomization: KustomizationConfig{
			Interval: 5 * time.Minute,
			Prune:    true,
			Health:   true,
		},
	}
}

// NetworkingCNI defines Container Network Interface configuration
type NetworkingCNI struct {
	Provider            string // flannel, calico, weave, cilium, etc
	CIDR                string
	ServiceCIDR         string
	ClusterDomain       string
	EnableNetworkPolicy bool
	EnableDualStack     bool
}

// NewNetworkingCNI creates a new CNI configuration
func NewNetworkingCNI(provider, cidr string) *NetworkingCNI {
	return &NetworkingCNI{
		Provider:            provider,
		CIDR:                cidr,
		ServiceCIDR:         "10.96.0.0/12",
		ClusterDomain:       "cluster.local",
		EnableNetworkPolicy: true,
		EnableDualStack:     false,
	}
}

// IsValidCIDR validates a CIDR block
func IsValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// StorageClass defines a Kubernetes storage class
type StorageClass struct {
	Name                 string
	Provisioner          string
	VolumeBindingMode    string // Immediate, WaitForFirstConsumer
	ReclaimPolicy        string // Delete, Retain, Recycle
	AllowVolumeExpansion bool
	Parameters           map[string]string
}

// NewStorageClass creates a new storage class
func NewStorageClass(name, provisioner string) *StorageClass {
	return &StorageClass{
		Name:                 name,
		Provisioner:          provisioner,
		VolumeBindingMode:    "Immediate",
		ReclaimPolicy:        "Delete",
		AllowVolumeExpansion: true,
		Parameters:           make(map[string]string),
	}
}

// GetCloudProvider returns appropriate cloud provider implementation
func GetCloudProvider(providerType string) (CloudProvider, error) {
	switch strings.ToLower(providerType) {
	case "aws":
		return &awsProvider{}, nil
	case "azure":
		return &azureProvider{}, nil
	case "gcp":
		return &gcpProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", providerType)
	}
}

type awsProvider struct{}

func (p *awsProvider) GetRegion() string                                      { return "us-east-1" }
func (p *awsProvider) GetEndpoint() string                                    { return "https://aws.amazon.com" }
func (p *awsProvider) Authenticate() error                                    { return nil }
func (p *awsProvider) ListResources(resourceType string) ([]string, error)    { return []string{}, nil }
func (p *awsProvider) CreateResource(resourceType, resourceName string) error { return nil }
func (p *awsProvider) DeleteResource(resourceType, resourceName string) error { return nil }

type azureProvider struct{}

func (p *azureProvider) GetRegion() string                                      { return "eastus" }
func (p *azureProvider) GetEndpoint() string                                    { return "https://azure.microsoft.com" }
func (p *azureProvider) Authenticate() error                                    { return nil }
func (p *azureProvider) ListResources(resourceType string) ([]string, error)    { return []string{}, nil }
func (p *azureProvider) CreateResource(resourceType, resourceName string) error { return nil }
func (p *azureProvider) DeleteResource(resourceType, resourceName string) error { return nil }

type gcpProvider struct{}

func (p *gcpProvider) GetRegion() string                                      { return "us-central1" }
func (p *gcpProvider) GetEndpoint() string                                    { return "https://cloud.google.com" }
func (p *gcpProvider) Authenticate() error                                    { return nil }
func (p *gcpProvider) ListResources(resourceType string) ([]string, error)    { return []string{}, nil }
func (p *gcpProvider) CreateResource(resourceType, resourceName string) error { return nil }
func (p *gcpProvider) DeleteResource(resourceType, resourceName string) error { return nil }
