package export

// =====================================================
// P2 resource-ification — Export.
//
// ExportJobResource wraps the imperative ExportJob struct so a
// controller can reconcile export operations as first-class
// platform resources.
// =====================================================

import (
	"time"

	"example.com/axiomnizam/internal/resources"
)

const (
	ExportJobKind       = "ExportJob"
	ExportJobAPIVersion = "export.axiomnizam.io/v1"
)

// ExportJobSpec is the desired state of an export job.
type ExportJobSpec struct {
	TenantID    string                 `json:"tenantId,omitempty"`
	Description string                 `json:"description,omitempty"`
	Format      ExportFormat           `json:"format"`
	Source      ExportSource           `json:"source"`
	Query       string                 `json:"query,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Columns     []string               `json:"columns,omitempty"`
	Compression CompressionType        `json:"compression,omitempty"`
	Encryption  EncryptionConfig       `json:"encryption,omitempty"`
	Destination ExportDestination      `json:"destination"`
	Schedule    ScheduleConfig         `json:"schedule,omitempty"`

	// Cancel, when true, asks the controller to cancel the export.
	Cancel bool `json:"cancel,omitempty"`
}

// ExportJobResourceStatus extends the canonical object status.
type ExportJobResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	ExportStatus  ExportStatus `json:"exportStatus"`
	Progress      float64      `json:"progress"`
	RecordCount   int64        `json:"recordCount"`
	FileSize      int64        `json:"fileSize"`
	ProcessedRows int64        `json:"processedRows"`
	SkippedRows   int64        `json:"skippedRows"`
	ErrorRows     int64        `json:"errorRows"`
	DownloadURL   string       `json:"downloadUrl,omitempty"`
	FailureReason string       `json:"failureReason,omitempty"`
	StartedAt     *time.Time   `json:"startedAt,omitempty"`
	CompletedAt   *time.Time   `json:"completedAt,omitempty"`
}

// ExportJobResource is the declarative resource for an ExportJob.
type ExportJobResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   ExportJobSpec           `json:"spec"`
	Status ExportJobResourceStatus `json:"status"`
}

func (r *ExportJobResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *ExportJobResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *ExportJobResource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *ExportJobResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *ExportJobResource) DeepCopy() resources.Resource {
	cp := *r
	if len(r.Spec.Columns) > 0 {
		cp.Spec.Columns = append([]string(nil), r.Spec.Columns...)
	}
	return &cp
}
func (r *ExportJobResource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *ExportJobResource) GetGeneration() int64         { return r.Generation }
func (r *ExportJobResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }
