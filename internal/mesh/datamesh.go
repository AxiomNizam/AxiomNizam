package mesh

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DataDomain represents a logical domain in the data mesh
type DataDomain struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Owner        string            `json:"owner"` // Domain owner team
	DataProducts []DataProduct     `json:"dataProducts"`
	Labels       map[string]string `json:"labels,omitempty"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

// DataProduct represents a data product within a domain
type DataProduct struct {
	Name          string                 `json:"name"`
	DomainName    string                 `json:"domainName"`
	Description   string                 `json:"description"`
	Owner         string                 `json:"owner"` // Team responsible
	SchemaVersion string                 `json:"schemaVersion"`
	Schema        map[string]interface{} `json:"schema"` // Data structure
	SLA           SLA                    `json:"sla"`
	Ports         []DataPort             `json:"ports"` // Output ports
	Subscriptions []Subscription         `json:"subscriptions"`
	Tags          []string               `json:"tags"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}

// DataPort is an output port of a data product
type DataPort struct {
	Name        string `json:"name"`
	Format      string `json:"format"`      // parquet, json, avro, etc.
	Location    string `json:"location"`    // S3, API endpoint, etc.
	Frequency   string `json:"frequency"`   // batch, realtime, etc.
	Credentials string `json:"credentials"` // Secret reference
}

// Subscription represents a subscription to a data product
type Subscription struct {
	ID            string            `json:"id"`
	SubscriberID  string            `json:"subscriberId"` // Domain/team ID
	DataProductID string            `json:"dataProductId"`
	PortName      string            `json:"portName"`
	Status        string            `json:"status"` // active, pending, cancelled
	CreatedAt     time.Time         `json:"createdAt"`
	AccessLevel   string            `json:"accessLevel"`
	Metadata      map[string]string `json:"metadata"`
}

// SLA defines service level agreements
type SLA struct {
	Availability    string `json:"availability"` // e.g., "99.9%"
	Latency         int    `json:"latency"`      // ms
	RPO             int    `json:"rpo"`          // Recovery Point Objective (seconds)
	RTO             int    `json:"rto"`          // Recovery Time Objective (seconds)
	SupportLevel    string `json:"supportLevel"` // P1, P2, P3
	UpdateFrequency string `json:"updateFrequency"`
}

// DataMesh manages the data mesh topology
type DataMesh struct {
	mu      sync.RWMutex
	domains map[string]*DataDomain
}

// NewDataMesh creates a new data mesh
func NewDataMesh() *DataMesh {
	return &DataMesh{
		domains: make(map[string]*DataDomain),
	}
}

// CreateDomain creates a new data domain
func (dm *DataMesh) CreateDomain(ctx context.Context, domain *DataDomain) error {
	if domain.Name == "" {
		return fmt.Errorf("domain name is required")
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	if _, exists := dm.domains[domain.Name]; exists {
		return fmt.Errorf("domain already exists: %s", domain.Name)
	}

	domain.CreatedAt = time.Now()
	domain.UpdatedAt = time.Now()
	dm.domains[domain.Name] = domain

	return nil
}

// GetDomain retrieves a domain
func (dm *DataMesh) GetDomain(name string) *DataDomain {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.domains[name]
}

// ListDomains lists all domains
func (dm *DataMesh) ListDomains() []*DataDomain {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	domains := make([]*DataDomain, 0, len(dm.domains))
	for _, d := range dm.domains {
		domains = append(domains, d)
	}
	return domains
}

// CreateDataProduct creates a data product in a domain
func (dm *DataMesh) CreateDataProduct(ctx context.Context, domainName string, product *DataProduct) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	domain, exists := dm.domains[domainName]
	if !exists {
		return fmt.Errorf("domain not found: %s", domainName)
	}

	if product.Name == "" {
		return fmt.Errorf("product name is required")
	}

	// Check for duplicates
	for _, existing := range domain.DataProducts {
		if existing.Name == product.Name {
			return fmt.Errorf("product already exists in domain: %s", product.Name)
		}
	}

	product.DomainName = domainName
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.Subscriptions = make([]Subscription, 0)

	domain.DataProducts = append(domain.DataProducts, *product)
	domain.UpdatedAt = time.Now()

	return nil
}

// GetDataProduct retrieves a data product
func (dm *DataMesh) GetDataProduct(domainName, productName string) *DataProduct {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	domain, exists := dm.domains[domainName]
	if !exists {
		return nil
	}

	for i, product := range domain.DataProducts {
		if product.Name == productName {
			return &domain.DataProducts[i]
		}
	}

	return nil
}

// Subscribe subscribes to a data product
func (dm *DataMesh) Subscribe(ctx context.Context, domainName, productName, subscriberID, portName string) (*Subscription, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	domain, exists := dm.domains[domainName]
	if !exists {
		return nil, fmt.Errorf("domain not found: %s", domainName)
	}

	var product *DataProduct
	for i := range domain.DataProducts {
		if domain.DataProducts[i].Name == productName {
			product = &domain.DataProducts[i]
			break
		}
	}

	if product == nil {
		return nil, fmt.Errorf("product not found: %s", productName)
	}

	// Create subscription
	subscription := Subscription{
		ID:            fmt.Sprintf("sub-%d", time.Now().Unix()),
		SubscriberID:  subscriberID,
		DataProductID: productName,
		PortName:      portName,
		Status:        "pending",
		CreatedAt:     time.Now(),
		AccessLevel:   "read",
	}

	product.Subscriptions = append(product.Subscriptions, subscription)
	domain.UpdatedAt = time.Now()

	return &subscription, nil
}

// UnsubscribeFromProduct unsubscribes from a data product
func (dm *DataMesh) Unsubscribe(ctx context.Context, subscriptionID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for _, domain := range dm.domains {
		for i := range domain.DataProducts {
			filtered := make([]Subscription, 0)
			found := false
			for _, sub := range domain.DataProducts[i].Subscriptions {
				if sub.ID != subscriptionID {
					filtered = append(filtered, sub)
				} else {
					found = true
				}
			}

			if found {
				domain.DataProducts[i].Subscriptions = filtered
				domain.UpdatedAt = time.Now()
				return nil
			}
		}
	}

	return fmt.Errorf("subscription not found: %s", subscriptionID)
}

// GetProductsByOwner gets data products owned by a team
func (dm *DataMesh) GetProductsByOwner(owner string) []DataProduct {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	products := make([]DataProduct, 0)
	for _, domain := range dm.domains {
		for _, product := range domain.DataProducts {
			if product.Owner == owner {
				products = append(products, product)
			}
		}
	}
	return products
}

// GetProductsByTag searches for products by tag
func (dm *DataMesh) GetProductsByTag(tag string) []DataProduct {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	products := make([]DataProduct, 0)
	for _, domain := range dm.domains {
		for _, product := range domain.DataProducts {
			for _, t := range product.Tags {
				if t == tag {
					products = append(products, product)
					break
				}
			}
		}
	}
	return products
}

// DataMeshDiscovery provides discovery and search capabilities
type DataMeshDiscovery struct {
	mesh *DataMesh
}

// NewDataMeshDiscovery creates a new discovery interface
func NewDataMeshDiscovery(mesh *DataMesh) *DataMeshDiscovery {
	return &DataMeshDiscovery{mesh: mesh}
}

// FindByTag finds products by tag
func (dmd *DataMeshDiscovery) FindByTag(tag string) []DataProduct {
	return dmd.mesh.GetProductsByTag(tag)
}

// FindByOwner finds products by owner
func (dmd *DataMeshDiscovery) FindByOwner(owner string) []DataProduct {
	return dmd.mesh.GetProductsByOwner(owner)
}

// FindByDomain finds all products in a domain
func (dmd *DataMeshDiscovery) FindByDomain(domainName string) []DataProduct {
	domain := dmd.mesh.GetDomain(domainName)
	if domain == nil {
		return []DataProduct{}
	}
	return domain.DataProducts
}

// LineageTracer tracks data lineage across the mesh
type LineageTracer struct {
	mesh *DataMesh
}

// NewLineageTracer creates a new lineage tracer
func NewLineageTracer(mesh *DataMesh) *LineageTracer {
	return &LineageTracer{mesh: mesh}
}

// TraceDownstream traces downstream consumers of a product
func (lt *LineageTracer) TraceDownstream(domainName, productName string) []Subscription {
	product := lt.mesh.GetDataProduct(domainName, productName)
	if product == nil {
		return []Subscription{}
	}
	return product.Subscriptions
}

// TraceUpstream traces upstream data sources (placeholder for future implementation)
func (lt *LineageTracer) TraceUpstream(domainName, productName string) []string {
	// Would trace input data sources
	return []string{}
}

// DiscoverRelated finds related data products
func (lt *LineageTracer) DiscoverRelated(domainName, productName string) []DataProduct {
	product := lt.mesh.GetDataProduct(domainName, productName)
	if product == nil {
		return []DataProduct{}
	}

	// Find products with overlapping tags or same owner
	related := make([]DataProduct, 0)
	for _, domain := range lt.mesh.ListDomains() {
		for _, other := range domain.DataProducts {
			if other.Name == productName && other.DomainName == domainName {
				continue // Skip self
			}

			// Check if same owner or overlapping tags
			if other.Owner == product.Owner {
				related = append(related, other)
				continue
			}

			for _, otherTag := range other.Tags {
				for _, myTag := range product.Tags {
					if otherTag == myTag {
						related = append(related, other)
						break
					}
				}
			}
		}
	}

	return related
}

// GlobalDataMesh is the package-level data mesh instance
var GlobalDataMesh = NewDataMesh()

// CreateDomain creates a domain via global mesh
func CreateDomain(ctx context.Context, domain *DataDomain) error {
	return GlobalDataMesh.CreateDomain(ctx, domain)
}

// GetDomain gets a domain via global mesh
func GetDomain(name string) *DataDomain {
	return GlobalDataMesh.GetDomain(name)
}

// ListDomains lists domains via global mesh
func ListDomains() []*DataDomain {
	return GlobalDataMesh.ListDomains()
}

// CreateDataProduct creates a product via global mesh
func CreateDataProduct(ctx context.Context, domainName string, product *DataProduct) error {
	return GlobalDataMesh.CreateDataProduct(ctx, domainName, product)
}

// Subscribe subscribes via global mesh
func Subscribe(ctx context.Context, domainName, productName, subscriberID, portName string) (*Subscription, error) {
	return GlobalDataMesh.Subscribe(ctx, domainName, productName, subscriberID, portName)
}
