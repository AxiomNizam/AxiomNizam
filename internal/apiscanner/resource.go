package apiscanner

// =====================================================
// P2 resource-ification — APIScan.
//
// APIScanResource is a declarative *scan request* resource.  The
// reconciler turns Spec (endpoint + options) into a ScanRequest,
// invokes the scanner, and records findings / summary on Status.
// This makes ad-hoc security scans a first-class platform resource
// with visible history, generation tracking, and retry semantics.
// =====================================================

import (
	"time"

	"example.com/axiomnizam/internal/resources"
)

const (
	APIScanKind       = "APIScan"
	APIScanAPIVersion = "apiscanner.axiomnizam.io/v1"
)

// APIScanSpec is the desired scan to execute.
type APIScanSpec struct {
	Endpoint           Endpoint      `json:"endpoint"`
	Timeout            time.Duration `json:"timeout,omitempty"`
	RetryCount         int           `json:"retryCount,omitempty"`
	RetryBackoff       time.Duration `json:"retryBackoff,omitempty"`
	InsecureSkipVerify bool          `json:"insecureSkipVerify,omitempty"`
	AuthHeader         string        `json:"authHeader,omitempty"`
	AuthValue          string        `json:"authValue,omitempty"`
	Format             OutputFormat  `json:"format,omitempty"`

	// RunOnce makes the scan execute a single time and then stay in
	// Completed state.  When false, the reconciler re-runs on a cadence
	// derived from timing.DefaultRequeueAfter.
	RunOnce bool `json:"runOnce,omitempty"`
}

// APIScanResourceStatus carries scan telemetry.
type APIScanResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	LastScanAt *time.Time  `json:"lastScanAt,omitempty"`
	LastResult *ScanResult `json:"lastResult,omitempty"`
	ScanCount  int         `json:"scanCount"`
	LastError  string      `json:"lastError,omitempty"`
}

// APIScanResource is the declarative resource for an APIScan job.
type APIScanResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   APIScanSpec           `json:"spec"`
	Status APIScanResourceStatus `json:"status"`
}

// --- resources.Resource implementation ---

func (r *APIScanResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *APIScanResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *APIScanResource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *APIScanResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *APIScanResource) DeepCopy() resources.Resource { cp := *r; return &cp }

// --- reconciler.Resource implementation ---

func (r *APIScanResource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *APIScanResource) GetGeneration() int64         { return r.Generation }
func (r *APIScanResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }
