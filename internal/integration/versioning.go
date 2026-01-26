package integration

import (
	"fmt"
)

// ========== API VERSION MANAGEMENT ==========

// APIVersionManager manages API versions and conversions
type APIVersionManager struct {
	versions         map[string]*APIVersionInfo
	preferredVersion string
}

// APIVersionInfo contains version metadata
type APIVersionInfo struct {
	Name           string
	Group          string
	Version        string
	Kind           string
	Plural         string
	Singular       string
	ShortNames     []string
	Categories     []string
	StorageVersion bool
	Served         bool
}

// NewAPIVersionManager creates manager
func NewAPIVersionManager() *APIVersionManager {
	avm := &APIVersionManager{
		versions: make(map[string]*APIVersionInfo),
	}

	// Register standard versions
	avm.RegisterVersion(&APIVersionInfo{
		Name:           "APIResource",
		Group:          "axiom.io",
		Version:        "v1alpha1",
		Kind:           "APIResource",
		Plural:         "apiresources",
		Singular:       "apiresource",
		ShortNames:     []string{"api"},
		Categories:     []string{"resources"},
		StorageVersion: false,
		Served:         true,
	})

	avm.RegisterVersion(&APIVersionInfo{
		Name:           "APIResource",
		Group:          "axiom.io",
		Version:        "v1beta1",
		Kind:           "APIResource",
		Plural:         "apiresources",
		Singular:       "apiresource",
		ShortNames:     []string{"api"},
		Categories:     []string{"resources"},
		StorageVersion: false,
		Served:         true,
	})

	avm.RegisterVersion(&APIVersionInfo{
		Name:           "APIResource",
		Group:          "axiom.io",
		Version:        "v1",
		Kind:           "APIResource",
		Plural:         "apiresources",
		Singular:       "apiresource",
		ShortNames:     []string{"api"},
		Categories:     []string{"resources"},
		StorageVersion: true,
		Served:         true,
	})

	avm.preferredVersion = "v1"

	return avm
}

// RegisterVersion registers API version
func (avm *APIVersionManager) RegisterVersion(info *APIVersionInfo) {
	key := fmt.Sprintf("%s/%s/%s", info.Group, info.Version, info.Kind)
	avm.versions[key] = info
}

// GetVersion gets version info
func (avm *APIVersionManager) GetVersion(group, version, kind string) *APIVersionInfo {
	key := fmt.Sprintf("%s/%s/%s", group, version, kind)
	return avm.versions[key]
}

// GetAllVersions gets all versions for kind
func (avm *APIVersionManager) GetAllVersions(kind string) []*APIVersionInfo {
	result := make([]*APIVersionInfo, 0)
	for _, info := range avm.versions {
		if info.Kind == kind {
			result = append(result, info)
		}
	}
	return result
}

// GetPreferredVersion gets preferred version for kind
func (avm *APIVersionManager) GetPreferredVersion(kind string) *APIVersionInfo {
	return avm.GetVersion("axiom.io", avm.preferredVersion, kind)
}

// SetPreferredVersion sets preferred version
func (avm *APIVersionManager) SetPreferredVersion(version string) {
	avm.preferredVersion = version
}

// CanConvert checks if conversion is supported
func (avm *APIVersionManager) CanConvert(fromVersion, toVersion string) bool {
	// Support conversions between v1alpha1, v1beta1, v1
	supportedVersions := map[string]bool{
		"v1alpha1": true,
		"v1beta1":  true,
		"v1":       true,
	}

	return supportedVersions[fromVersion] && supportedVersions[toVersion]
}

// ========== RESOURCE CONVERSION STRATEGIES ==========

// ConversionStrategy defines how to convert between versions
type ConversionStrategy interface {
	Convert(obj map[string]interface{}) (map[string]interface{}, error)
}

// V1Alpha1ToV1Strategy converts v1alpha1 to v1
type V1Alpha1ToV1Strategy struct{}

// Convert implements ConversionStrategy
func (vs *V1Alpha1ToV1Strategy) Convert(obj map[string]interface{}) (map[string]interface{}, error) {
	// Copy object
	result := make(map[string]interface{})
	for k, v := range obj {
		result[k] = v
	}

	// Update API version
	if meta, ok := obj["metadata"].(map[string]interface{}); ok {
		newMeta := make(map[string]interface{})
		for k, v := range meta {
			newMeta[k] = v
		}
		newMeta["apiVersion"] = "axiom.io/v1"
		result["metadata"] = newMeta
	}

	return result, nil
}

// V1BetaToV1Strategy converts v1beta1 to v1
type V1BetaToV1Strategy struct{}

// Convert implements ConversionStrategy
func (vs *V1BetaToV1Strategy) Convert(obj map[string]interface{}) (map[string]interface{}, error) {
	// Copy object
	result := make(map[string]interface{})
	for k, v := range obj {
		result[k] = v
	}

	// Update API version
	if meta, ok := obj["metadata"].(map[string]interface{}); ok {
		newMeta := make(map[string]interface{})
		for k, v := range meta {
			newMeta[k] = v
		}
		newMeta["apiVersion"] = "axiom.io/v1"
		result["metadata"] = newMeta
	}

	return result, nil
}

// ========== CONVERSION REGISTRY ==========

// ConversionRegistry manages conversion strategies
type ConversionRegistry struct {
	strategies map[string]ConversionStrategy
}

// NewConversionRegistry creates registry
func NewConversionRegistry() *ConversionRegistry {
	cr := &ConversionRegistry{
		strategies: make(map[string]ConversionStrategy),
	}

	// Register standard conversions
	cr.RegisterConversion("v1alpha1->v1", &V1Alpha1ToV1Strategy{})
	cr.RegisterConversion("v1beta1->v1", &V1BetaToV1Strategy{})

	return cr
}

// RegisterConversion registers conversion strategy
func (cr *ConversionRegistry) RegisterConversion(key string, strategy ConversionStrategy) {
	cr.strategies[key] = strategy
}

// GetConversion gets conversion strategy
func (cr *ConversionRegistry) GetConversion(fromVersion, toVersion string) ConversionStrategy {
	key := fmt.Sprintf("%s->%s", fromVersion, toVersion)
	return cr.strategies[key]
}

// Convert converts object between versions
func (cr *ConversionRegistry) Convert(fromVersion, toVersion string, obj map[string]interface{}) (map[string]interface{}, error) {
	if fromVersion == toVersion {
		return obj, nil
	}

	strategy := cr.GetConversion(fromVersion, toVersion)
	if strategy == nil {
		return nil, fmt.Errorf("no conversion strategy from %s to %s", fromVersion, toVersion)
	}

	return strategy.Convert(obj)
}

// ========== DEPRECATION MANAGER ==========

// DeprecationManager tracks deprecated API versions
type DeprecationManager struct {
	deprecated map[string]*DeprecationInfo
}

// DeprecationInfo contains deprecation details
type DeprecationInfo struct {
	Version      string
	DeprecatedIn string
	RemovedIn    string
	Replacement  string
	Warning      string
}

// NewDeprecationManager creates manager
func NewDeprecationManager() *DeprecationManager {
	dm := &DeprecationManager{
		deprecated: make(map[string]*DeprecationInfo),
	}

	// Register deprecations
	dm.MarkDeprecated("v1alpha1", &DeprecationInfo{
		Version:      "v1alpha1",
		DeprecatedIn: "axiom-1.0",
		RemovedIn:    "axiom-2.0",
		Replacement:  "v1",
		Warning:      "v1alpha1 is deprecated, please migrate to v1",
	})

	dm.MarkDeprecated("v1beta1", &DeprecationInfo{
		Version:      "v1beta1",
		DeprecatedIn: "axiom-1.5",
		RemovedIn:    "axiom-2.0",
		Replacement:  "v1",
		Warning:      "v1beta1 is deprecated, please migrate to v1",
	})

	return dm
}

// MarkDeprecated marks version as deprecated
func (dm *DeprecationManager) MarkDeprecated(version string, info *DeprecationInfo) {
	dm.deprecated[version] = info
}

// IsDeprecated checks if version is deprecated
func (dm *DeprecationManager) IsDeprecated(version string) bool {
	_, ok := dm.deprecated[version]
	return ok
}

// GetDeprecationInfo gets deprecation info
func (dm *DeprecationManager) GetDeprecationInfo(version string) *DeprecationInfo {
	return dm.deprecated[version]
}

// GetWarning gets deprecation warning
func (dm *DeprecationManager) GetWarning(version string) string {
	info := dm.GetDeprecationInfo(version)
	if info != nil {
		return info.Warning
	}
	return ""
}

// ========== API GROUP MANAGER ==========

// APIGroupManager manages API groups
type APIGroupManager struct {
	groups map[string]*APIGroup
}

// APIGroup represents an API group
type APIGroup struct {
	Name             string
	Versions         []*GroupVersion
	PreferredVersion *GroupVersion
}

// GroupVersion represents version in a group
type GroupVersion struct {
	GroupVersion string // "axiom.io/v1"
	Version      string // "v1"
}

// NewAPIGroupManager creates manager
func NewAPIGroupManager() *APIGroupManager {
	agm := &APIGroupManager{
		groups: make(map[string]*APIGroup),
	}

	// Register axiom.io group
	group := &APIGroup{
		Name: "axiom.io",
		Versions: []*GroupVersion{
			{GroupVersion: "axiom.io/v1alpha1", Version: "v1alpha1"},
			{GroupVersion: "axiom.io/v1beta1", Version: "v1beta1"},
			{GroupVersion: "axiom.io/v1", Version: "v1"},
		},
		PreferredVersion: &GroupVersion{GroupVersion: "axiom.io/v1", Version: "v1"},
	}
	agm.groups["axiom.io"] = group

	return agm
}

// GetGroup gets API group
func (agm *APIGroupManager) GetGroup(name string) *APIGroup {
	return agm.groups[name]
}

// ListGroups lists all API groups
func (agm *APIGroupManager) ListGroups() []*APIGroup {
	result := make([]*APIGroup, 0, len(agm.groups))
	for _, group := range agm.groups {
		result = append(result, group)
	}
	return result
}

// RegisterGroup registers API group
func (agm *APIGroupManager) RegisterGroup(group *APIGroup) {
	agm.groups[group.Name] = group
}
