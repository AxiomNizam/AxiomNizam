package utils

import (
	"fmt"
	"time"
)

// ContainerConfig holds container-specific configuration
type ContainerConfig struct {
	ImageName   string
	ImageTag    string
	ContainerID string
	Namespace   string
	PodName     string
	NodeName    string
	Labels      map[string]string
	Annotations map[string]string
}

// KubernetesMetadata holds Kubernetes-specific metadata
type KubernetesMetadata struct {
	ClusterName    string
	Namespace      string
	PodName        string
	ContainerName  string
	NodeName       string
	ServiceName    string
	Version        string
	DeploymentName string
}

// NewKubernetesMetadata creates new Kubernetes metadata
func NewKubernetesMetadata() *KubernetesMetadata {
	return &KubernetesMetadata{
		Version: "1.0.0",
	}
}

// ServiceRegistry manages service discovery
type ServiceRegistry struct {
	services map[string]ServiceInfo
}

// ServiceInfo holds service information
type ServiceInfo struct {
	Name     string
	Address  string
	Port     int
	Protocol string
	Health   bool
	Tags     []string
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]ServiceInfo),
	}
}

// RegisterService registers a service in the registry
func (sr *ServiceRegistry) RegisterService(service ServiceInfo) error {
	if service.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	sr.services[service.Name] = service
	return nil
}

// DeregisterService removes a service from the registry
func (sr *ServiceRegistry) DeregisterService(name string) error {
	if _, exists := sr.services[name]; !exists {
		return fmt.Errorf("service not found: %s", name)
	}
	delete(sr.services, name)
	return nil
}

// GetService retrieves a service from the registry
func (sr *ServiceRegistry) GetService(name string) (ServiceInfo, error) {
	service, exists := sr.services[name]
	if !exists {
		return ServiceInfo{}, fmt.Errorf("service not found: %s", name)
	}
	return service, nil
}

// ListServices returns all registered services
func (sr *ServiceRegistry) ListServices() map[string]ServiceInfo {
	return sr.services
}

// UpdateServiceHealth updates the health status of a service
func (sr *ServiceRegistry) UpdateServiceHealth(name string, healthy bool) error {
	service, exists := sr.services[name]
	if !exists {
		return fmt.Errorf("service not found: %s", name)
	}
	service.Health = healthy
	sr.services[name] = service
	return nil
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name         string
	failureCount int
	successCount int
	lastFailTime time.Time
	state        string // "closed", "open", "half-open"
	threshold    int
	timeout      time.Duration
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:      name,
		state:     "closed",
		threshold: threshold,
		timeout:   timeout,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	if cb.state == "open" {
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = "half-open"
			cb.successCount = 0
		} else {
			return fmt.Errorf("circuit breaker is open for %s", cb.name)
		}
	}

	err := fn()

	if err != nil {
		cb.failureCount++
		cb.lastFailTime = time.Now()
		if cb.failureCount >= cb.threshold {
			cb.state = "open"
		}
		return err
	}

	cb.failureCount = 0
	if cb.state == "half-open" {
		cb.successCount++
		if cb.successCount >= 2 {
			cb.state = "closed"
			cb.successCount = 0
		}
	}

	return nil
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() string {
	return cb.state
}

// ServiceMesh provides service mesh integration utilities
type ServiceMesh struct {
	version string
	config  map[string]interface{}
}

// NewServiceMesh creates a new service mesh instance
func NewServiceMesh(version string) *ServiceMesh {
	return &ServiceMesh{
		version: version,
		config:  make(map[string]interface{}),
	}
}

// SetConfig sets a configuration value
func (sm *ServiceMesh) SetConfig(key string, value interface{}) {
	sm.config[key] = value
}

// GetConfig retrieves a configuration value
func (sm *ServiceMesh) GetConfig(key string) (interface{}, bool) {
	val, exists := sm.config[key]
	return val, exists
}

// VirtualService defines a virtual service for traffic management
type VirtualService struct {
	Name         string
	Hosts        []string
	Routes       []Route
	Retries      *RetryPolicy
	Timeout      time.Duration
	CircuitBreak *CircuitBreakerPolicy
}

// Route defines a traffic route
type Route struct {
	Match       RouteMatch
	Destination RouteDestination
	Weight      int
}

// RouteMatch defines matching criteria for a route
type RouteMatch struct {
	URI         string
	Method      string
	Headers     map[string]string
	QueryParams map[string]string
}

// RouteDestination defines the destination for a route
type RouteDestination struct {
	Host   string
	Port   int
	Subset string
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	Attempts      int
	PerTryTimeout time.Duration
	RetryOn       []string
}

// CircuitBreakerPolicy defines circuit breaker behavior
type CircuitBreakerPolicy struct {
	ConsecutiveErrors int
	Interval          time.Duration
	MaxRequests       int
}

// LoadBalancer manages load balancing strategies
type LoadBalancer struct {
	strategy string
	services []ServiceInfo
	current  int
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(strategy string) *LoadBalancer {
	return &LoadBalancer{
		strategy: strategy,
		services: make([]ServiceInfo, 0),
		current:  0,
	}
}

// AddService adds a service to the load balancer
func (lb *LoadBalancer) AddService(service ServiceInfo) {
	lb.services = append(lb.services, service)
}

// GetNextService returns the next service based on the load balancing strategy
func (lb *LoadBalancer) GetNextService() (ServiceInfo, error) {
	if len(lb.services) == 0 {
		return ServiceInfo{}, fmt.Errorf("no services available")
	}

	switch lb.strategy {
	case "round-robin":
		service := lb.services[lb.current]
		lb.current = (lb.current + 1) % len(lb.services)
		return service, nil
	case "random":
		// In production, use actual random selection
		return lb.services[0], nil
	case "least-connections":
		// In production, track active connections
		return lb.services[0], nil
	default:
		return lb.services[0], nil
	}
}

// DistributedTracing provides distributed tracing utilities
type DistributedTracing struct {
	serviceName  string
	traceID      string
	spanID       string
	parentSpanID string
	samplingRate float64
}

// NewDistributedTracing creates a new distributed tracing instance
func NewDistributedTracing(serviceName string) *DistributedTracing {
	return &DistributedTracing{
		serviceName:  serviceName,
		samplingRate: 0.1,
	}
}

// GenerateTraceID generates a new trace ID
func (dt *DistributedTracing) GenerateTraceID() string {
	helper := NewCryptographicHelper()
	traceID, _ := helper.GenerateSecureToken(16)
	dt.traceID = traceID
	return traceID
}

// GenerateSpanID generates a new span ID
func (dt *DistributedTracing) GenerateSpanID() string {
	helper := NewCryptographicHelper()
	spanID, _ := helper.GenerateSecureToken(8)
	dt.spanID = spanID
	return spanID
}

// GetTraceContext returns the current trace context
func (dt *DistributedTracing) GetTraceContext() map[string]string {
	return map[string]string{
		"trace-id":       dt.traceID,
		"span-id":        dt.spanID,
		"parent-span-id": dt.parentSpanID,
		"service-name":   dt.serviceName,
		"sampling-rate":  fmt.Sprintf("%f", dt.samplingRate),
	}
}

// PodDisruptionBudget manages pod disruption
type PodDisruptionBudget struct {
	Name                       string
	Namespace                  string
	MinAvailable               int
	MaxUnavailable             int
	SelectorLabels             map[string]string
	UnhealthyPodEvictionPolicy string
}

// NewPodDisruptionBudget creates a new PDB
func NewPodDisruptionBudget(name, namespace string) *PodDisruptionBudget {
	return &PodDisruptionBudget{
		Name:           name,
		Namespace:      namespace,
		MinAvailable:   1,
		SelectorLabels: make(map[string]string),
	}
}

// NetworkPolicy manages network policies
type NetworkPolicy struct {
	Name        string
	Namespace   string
	Ingress     []NetworkPolicyRule
	Egress      []NetworkPolicyRule
	PodSelector map[string]string
}

// NetworkPolicyRule defines a network policy rule
type NetworkPolicyRule struct {
	Protocol string
	Port     int
	From     []string
	To       []string
}

// NewNetworkPolicy creates a new network policy
func NewNetworkPolicy(name, namespace string) *NetworkPolicy {
	return &NetworkPolicy{
		Name:        name,
		Namespace:   namespace,
		Ingress:     make([]NetworkPolicyRule, 0),
		Egress:      make([]NetworkPolicyRule, 0),
		PodSelector: make(map[string]string),
	}
}

// ResourceQuota manages resource quotas
type ResourceQuota struct {
	Name      string
	Namespace string
	Requests  ResourceRequirements
	Limits    ResourceRequirements
	PodCount  int
	PVCCount  int
}

// ResourceRequirements defines resource requirements
type ResourceRequirements struct {
	CPU    string
	Memory string
}

// NewResourceQuota creates a new resource quota
func NewResourceQuota(name, namespace string) *ResourceQuota {
	return &ResourceQuota{
		Name:      name,
		Namespace: namespace,
	}
}

// ServiceLevelObjective defines SLO targets
type ServiceLevelObjective struct {
	Name           string
	Target         float64 // percentage
	Window         time.Duration
	ErrorBudget    float64
	AlertThreshold float64
}

// NewServiceLevelObjective creates a new SLO
func NewServiceLevelObjective(name string, target float64) *ServiceLevelObjective {
	return &ServiceLevelObjective{
		Name:           name,
		Target:         target,
		Window:         30 * 24 * time.Hour,
		ErrorBudget:    (100 - target) / 100,
		AlertThreshold: (100 - (target - 5)) / 100,
	}
}

// CanaryDeployment manages canary deployment strategy
type CanaryDeployment struct {
	Name          string
	StableVersion string
	CanaryVersion string
	CanaryWeight  int // 0-100
	ErrorRate     float64
	LatencyMetric time.Duration
}

// NewCanaryDeployment creates a new canary deployment
func NewCanaryDeployment(name string, stable, canary string) *CanaryDeployment {
	return &CanaryDeployment{
		Name:          name,
		StableVersion: stable,
		CanaryVersion: canary,
		CanaryWeight:  10,
		ErrorRate:     0.01,
		LatencyMetric: 500 * time.Millisecond,
	}
}

// ShouldPromoteCanary determines if canary should be promoted to stable
func (cd *CanaryDeployment) ShouldPromoteCanary(currentErrorRate float64, currentLatency time.Duration) bool {
	return currentErrorRate < cd.ErrorRate && currentLatency < cd.LatencyMetric
}
