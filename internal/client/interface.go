package client

import (
	"context"
)

// Resource represents a generic AxiomNizam resource
type Resource struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   ResourceMetadata       `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
	Status     map[string]interface{} `json:"status,omitempty"`
}

// ResourceMetadata contains resource metadata
type ResourceMetadata struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Generation int64             `json:"generation,omitempty"`
}

// ListOptions controls list operations
type ListOptions struct {
	Kind      string
	Namespace string
	Labels    map[string]string
}

// ResourceClient defines the interface for interacting with AxiomNizam resources
// via the API server.
//
// CRITICAL PRINCIPLE: All CLI operations go through this interface.
// There is NO direct database access from the CLI.
//
// Flow: CLI → ResourceClient → API Server → Controllers → Store
type ResourceClient interface {
	// Get retrieves a single resource by kind and name
	// Returns error if resource not found
	Get(ctx context.Context, kind, name string) (*Resource, error)

	// List retrieves resources of a given kind
	// Returns empty slice if no resources found
	List(ctx context.Context, opts ListOptions) ([]Resource, error)

	// Apply creates or updates a resource
	// - If resource doesn't exist: creates it and enqueues for reconciliation
	// - If resource exists: updates spec and enqueues for reconciliation
	// Returns the created/updated resource
	Apply(ctx context.Context, resource *Resource) (*Resource, error)

	// Delete removes a resource
	// Returns error if resource not found
	Delete(ctx context.Context, kind, name string) error

	// Watch watches for changes to resources (async)
	Watch(ctx context.Context, kind string) (<-chan *Resource, error)

	// GetStatus retrieves the status of a resource
	GetStatus(ctx context.Context, kind, name string) (map[string]interface{}, error)
}

// NewResourceClient creates a ResourceClient from an HTTP client
// This is the only way the CLI should communicate with the backend
func NewResourceClient(httpClient *Client) ResourceClient {
	return &apiResourceClient{
		client: httpClient,
	}
}

// apiResourceClient implements ResourceClient using HTTP
type apiResourceClient struct {
	client *Client
}

// Get retrieves a single resource
func (rc *apiResourceClient) Get(ctx context.Context, kind, name string) (*Resource, error) {
	resp, err := rc.client.Get(ctx, "/api/v1/"+kind+"/"+name, nil)
	if err != nil {
		return nil, err
	}

	var resource Resource
	if err := resp.ParseJSON(&resource); err != nil {
		return nil, err
	}

	return &resource, nil
}

// List retrieves resources
func (rc *apiResourceClient) List(ctx context.Context, opts ListOptions) ([]Resource, error) {
	query := make(map[string][]string)
	if opts.Namespace != "" {
		query["namespace"] = []string{opts.Namespace}
	}

	resp, err := rc.client.Get(ctx, "/api/v1/"+opts.Kind, nil)
	if err != nil {
		return nil, err
	}

	var resources []Resource
	if err := resp.ParseJSON(&resources); err != nil {
		return nil, err
	}

	return resources, nil
}

// Apply creates or updates a resource
func (rc *apiResourceClient) Apply(ctx context.Context, resource *Resource) (*Resource, error) {
	resp, err := rc.client.Post(ctx, "/api/v1/"+resource.Kind+"/apply", resource)
	if err != nil {
		return nil, err
	}

	var result Resource
	if err := resp.ParseJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete removes a resource
func (rc *apiResourceClient) Delete(ctx context.Context, kind, name string) error {
	_, err := rc.client.Delete(ctx, "/api/v1/"+kind+"/"+name)
	return err
}

// Watch watches for changes (stub - full implementation in API server)
func (rc *apiResourceClient) Watch(ctx context.Context, kind string) (<-chan *Resource, error) {
	// This would require WebSocket support - stubbed for now
	ch := make(<-chan *Resource)
	return ch, nil
}

// GetStatus retrieves resource status
func (rc *apiResourceClient) GetStatus(ctx context.Context, kind, name string) (map[string]interface{}, error) {
	resp, err := rc.client.Get(ctx, "/api/v1/"+kind+"/"+name+"/status", nil)
	if err != nil {
		return nil, err
	}

	var status map[string]interface{}
	if err := resp.ParseJSON(&status); err != nil {
		return nil, err
	}

	return status, nil
}
