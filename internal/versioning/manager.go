package versioning

import (
	"fmt"
	"sync"
	"time"
)

// APIVersion represents an API version
type APIVersion struct {
	Version          string
	Title            string
	Description      string
	Endpoints        map[string]*VersionedEndpoint
	DeprecationDate  *time.Time
	SunsetDate       *time.Time
	Status           string // active, deprecated, sunset
	BreakingChanges  []string
	MigrationGuide   string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// VersionedEndpoint represents an endpoint in a specific version
type VersionedEndpoint struct {
	Path             string
	Method           string
	Version          string
	RequestSchema    map[string]interface{}
	ResponseSchema   map[string]interface{}
	Deprecated       bool
	DeprecatedReason string
	ReplacedBy       string
	Transformer      func(interface{}) interface{}
}

// APIVersionManager manages API versions and transformations
type APIVersionManager struct {
	mu                 sync.RWMutex
	versions           map[string]*APIVersion
	currentVersion     string
	defaultVersion     string
	transformers       map[string]map[string]func(interface{}) interface{}
	versionUsage       map[string]int64
	requestLog         []*VersionedRequest
	maxRequestLogSize  int
	deprecationPolicy  *DeprecationPolicy
	endpointMigration  map[string]*MigrationPath
}

// DeprecationPolicy defines version deprecation rules
type DeprecationPolicy struct {
	NoticeBeforeSunset time.Duration
	MinSupportPeriod   time.Duration
	MajorVersionGap    int
}

// MigrationPath defines migration steps from one version to another
type MigrationPath struct {
	FromVersion string
	ToVersion   string
	Steps       []*MigrationStep
	Automatic   bool
}

// MigrationStep represents a migration action
type MigrationStep struct {
	Step        int
	Description string
	Action      string // rename_field, transform_value, add_field, remove_field
	Mapping     map[string]interface{}
}

// VersionedRequest logs a versioned API request
type VersionedRequest struct {
	ID        string
	Timestamp time.Time
	ClientID  string
	Version   string
	Endpoint  string
	Method    string
	Deprecated bool
}

// NewAPIVersionManager creates a new version manager
func NewAPIVersionManager(defaultVersion string) *APIVersionManager {
	return &APIVersionManager{
		versions:          make(map[string]*APIVersion),
		currentVersion:    defaultVersion,
		defaultVersion:    defaultVersion,
		transformers:      make(map[string]map[string]func(interface{}) interface{}),
		versionUsage:      make(map[string]int64),
		requestLog:        make([]*VersionedRequest, 0),
		maxRequestLogSize: 10000,
		endpointMigration: make(map[string]*MigrationPath),
		deprecationPolicy: &DeprecationPolicy{
			NoticeBeforeSunset: 90 * 24 * time.Hour,
			MinSupportPeriod:   6 * 24 * time.Hour,
			MajorVersionGap:    2,
		},
	}
}

// RegisterVersion registers a new API version
func (avm *APIVersionManager) RegisterVersion(version *APIVersion) error {
	avm.mu.Lock()
	defer avm.mu.Unlock()

	if _, exists := avm.versions[version.Version]; exists {
		return fmt.Errorf("version %s already registered", version.Version)
	}

	version.CreatedAt = time.Now()
	version.UpdatedAt = time.Now()

	avm.versions[version.Version] = version
	avm.versionUsage[version.Version] = 0

	return nil
}

// GetVersion retrieves a version
func (avm *APIVersionManager) GetVersion(version string) (*APIVersion, error) {
	avm.mu.RLock()
	defer avm.mu.RUnlock()

	v, exists := avm.versions[version]
	if !exists {
		return nil, fmt.Errorf("version %s not found", version)
	}

	return v, nil
}

// RegisterEndpoint registers an endpoint for a version
func (avm *APIVersionManager) RegisterEndpoint(version, path, method string, endpoint *VersionedEndpoint) error {
	avm.mu.Lock()
	defer avm.mu.Unlock()

	v, exists := avm.versions[version]
	if !exists {
		return fmt.Errorf("version %s not found", version)
	}

	endpoint.Version = version
	endpoint.Path = path
	endpoint.Method = method

	key := fmt.Sprintf("%s-%s", method, path)
	v.Endpoints[key] = endpoint

	return nil
}

// DeprecateEndpoint marks an endpoint as deprecated
func (avm *APIVersionManager) DeprecateEndpoint(version, path, method, reason, replacedBy string) error {
	avm.mu.Lock()
	defer avm.mu.Unlock()

	v, exists := avm.versions[version]
	if !exists {
		return fmt.Errorf("version %s not found", version)
	}

	key := fmt.Sprintf("%s-%s", method, path)
	endpoint, exists := v.Endpoints[key]
	if !exists {
		return fmt.Errorf("endpoint not found")
	}

	endpoint.Deprecated = true
	endpoint.DeprecatedReason = reason
	endpoint.ReplacedBy = replacedBy

	return nil
}

// DeprecateVersion marks an entire API version as deprecated
func (avm *APIVersionManager) DeprecateVersion(version, reason string) error {
	avm.mu.Lock()
	defer avm.mu.Unlock()

	v, exists := avm.versions[version]
	if !exists {
		return fmt.Errorf("version %s not found", version)
	}

	now := time.Now()
	sunsetDate := now.Add(avm.deprecationPolicy.NoticeBeforeSunset)

	v.Status = "deprecated"
	v.DeprecationDate = &now
	v.SunsetDate = &sunsetDate

	for _, endpoint := range v.Endpoints {
		endpoint.Deprecated = true
		endpoint.DeprecatedReason = reason
	}

	return nil
}

// RegisterTransformer registers data transformer between versions
func (avm *APIVersionManager) RegisterTransformer(fromVersion, toVersion string, transformer func(interface{}) interface{}) {
	avm.mu.Lock()
	defer avm.mu.Unlock()

	key := fmt.Sprintf("%s->%s", fromVersion, toVersion)

	if _, exists := avm.transformers[key]; !exists {
		avm.transformers[key] = make(map[string]func(interface{}) interface{})
	}

	avm.transformers[key]["transform"] = transformer
}

// TransformRequest transforms request between versions
func (avm *APIVersionManager) TransformRequest(data interface{}, fromVersion, toVersion string) (interface{}, error) {
	avm.mu.RLock()
	key := fmt.Sprintf("%s->%s", fromVersion, toVersion)
	transformer, exists := avm.transformers[key]
	avm.mu.RUnlock()

	if !exists || transformer["transform"] == nil {
		return data, nil
	}

	return transformer["transform"](data), nil
}

// TransformResponse transforms response between versions
func (avm *APIVersionManager) TransformResponse(data interface{}, fromVersion, toVersion string) (interface{}, error) {
	return avm.TransformRequest(data, fromVersion, toVersion)
}

// LogRequest logs a versioned request
func (avm *APIVersionManager) LogRequest(clientID, version, endpoint, method string) {
	avm.mu.Lock()
	defer avm.mu.Unlock()

	req := &VersionedRequest{
		ID:        fmt.Sprintf("req-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		ClientID:  clientID,
		Version:   version,
		Endpoint:  endpoint,
		Method:    method,
		Deprecated: false,
	}

	// Check if endpoint is deprecated
	if v, exists := avm.versions[version]; exists {
		key := fmt.Sprintf("%s-%s", method, endpoint)
		if ep, exists := v.Endpoints[key]; exists {
			req.Deprecated = ep.Deprecated
		}
	}

	avm.requestLog = append(avm.requestLog, req)
	avm.versionUsage[version]++

	if len(avm.requestLog) > avm.maxRequestLogSize {
		avm.requestLog = avm.requestLog[1:]
	}
}

// GetVersionUsage returns usage statistics for versions
func (avm *APIVersionManager) GetVersionUsage() map[string]int64 {
	avm.mu.RLock()
	defer avm.mu.RUnlock()

	usage := make(map[string]int64)
	for k, v := range avm.versionUsage {
		usage[k] = v
	}

	return usage
}

// GetDeprecationWarnings returns deprecation warnings
func (avm *APIVersionManager) GetDeprecationWarnings(version string) []string {
	avm.mu.RLock()
	defer avm.mu.RUnlock()

	warnings := make([]string, 0)

	v, exists := avm.versions[version]
	if !exists {
		return warnings
	}

	// Version deprecation warning
	if v.Status == "deprecated" && v.SunsetDate != nil {
		warnings = append(warnings, fmt.Sprintf("Version %s is deprecated and will sunset on %s", version, v.SunsetDate.Format("2006-01-02")))
	}

	// Endpoint deprecation warnings
	for _, endpoint := range v.Endpoints {
		if endpoint.Deprecated {
			msg := fmt.Sprintf("Endpoint %s %s is deprecated", endpoint.Method, endpoint.Path)
			if endpoint.ReplacedBy != "" {
				msg += fmt.Sprintf(" - use %s instead", endpoint.ReplacedBy)
			}
			warnings = append(warnings, msg)
		}
	}

	return warnings
}

// CreateMigrationPath creates a migration path
func (avm *APIVersionManager) CreateMigrationPath(fromVersion, toVersion string) (*MigrationPath, error) {
	avm.mu.Lock()
	defer avm.mu.Unlock()

	path := &MigrationPath{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Steps:       make([]*MigrationStep, 0),
		Automatic:   false,
	}

	key := fmt.Sprintf("%s->%s", fromVersion, toVersion)
	avm.endpointMigration[key] = path

	return path, nil
}

// AddMigrationStep adds a step to migration path
func (avm *APIVersionManager) AddMigrationStep(fromVersion, toVersion string, step *MigrationStep) error {
	avm.mu.Lock()
	defer avm.mu.Unlock()

	key := fmt.Sprintf("%s->%s", fromVersion, toVersion)
	path, exists := avm.endpointMigration[key]
	if !exists {
		return fmt.Errorf("migration path not found")
	}

	step.Step = len(path.Steps) + 1
	path.Steps = append(path.Steps, step)

	return nil
}

// GetMigrationGuide returns migration guide
func (avm *APIVersionManager) GetMigrationGuide(fromVersion, toVersion string) (string, error) {
	avm.mu.RLock()
	defer avm.mu.RUnlock()

	key := fmt.Sprintf("%s->%s", fromVersion, toVersion)
	path, exists := avm.endpointMigration[key]
	if !exists {
		return "", fmt.Errorf("migration path not found")
	}

	guide := fmt.Sprintf("Migration Guide: %s → %s\n\n", fromVersion, toVersion)

	for _, step := range path.Steps {
		guide += fmt.Sprintf("Step %d: %s\n", step.Step, step.Description)
		guide += fmt.Sprintf("  Action: %s\n", step.Action)
		if len(step.Mapping) > 0 {
			guide += fmt.Sprintf("  Mapping: %v\n", step.Mapping)
		}
	}

	return guide, nil
}

// ListAvailableVersions lists all available versions
func (avm *APIVersionManager) ListAvailableVersions() []string {
	avm.mu.RLock()
	defer avm.mu.RUnlock()

	versions := make([]string, 0)
	for v := range avm.versions {
		versions = append(versions, v)
	}

	return versions
}

// GetVersionInfo returns comprehensive version information
func (avm *APIVersionManager) GetVersionInfo() map[string]interface{} {
	avm.mu.RLock()
	defer avm.mu.RUnlock()

	info := make(map[string]interface{})
	info["current_version"] = avm.currentVersion
	info["default_version"] = avm.defaultVersion

	versions := make([]map[string]interface{}, 0)
	for vname, v := range avm.versions {
		vinfo := map[string]interface{}{
			"name":            vname,
			"title":           v.Title,
			"status":          v.Status,
			"endpoint_count":  len(v.Endpoints),
			"usage_count":     avm.versionUsage[vname],
			"created_at":      v.CreatedAt,
			"updated_at":      v.UpdatedAt,
		}

		if v.DeprecationDate != nil {
			vinfo["deprecation_date"] = v.DeprecationDate
		}
		if v.SunsetDate != nil {
			vinfo["sunset_date"] = v.SunsetDate
		}

		versions = append(versions, vinfo)
	}

	info["versions"] = versions
	info["total_versions"] = len(avm.versions)

	return info
}

// GetEndpointHistory returns endpoint changes across versions
func (avm *APIVersionManager) GetEndpointHistory(path, method string) []map[string]interface{} {
	avm.mu.RLock()
	defer avm.mu.RUnlock()

	history := make([]map[string]interface{}, 0)

	key := fmt.Sprintf("%s-%s", method, path)

	for vname, v := range avm.versions {
		if endpoint, exists := v.Endpoints[key]; exists {
			entry := map[string]interface{}{
				"version":   vname,
				"path":      endpoint.Path,
				"method":    endpoint.Method,
				"deprecated": endpoint.Deprecated,
			}

			if endpoint.Deprecated {
				entry["deprecation_reason"] = endpoint.DeprecatedReason
				entry["replaced_by"] = endpoint.ReplacedBy
			}

			history = append(history, entry)
		}
	}

	return history
}
