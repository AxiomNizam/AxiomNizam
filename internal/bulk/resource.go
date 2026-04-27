package bulk

// =====================================================
// P2 resource-ification — BulkOperation.
//
// BulkOperationResource wraps the imperative BulkOperation struct
// so a controller can reconcile bulk operations as first-class
// platform resources with Spec/Status separation.
// =====================================================

import (
	"time"

	"example.com/axiomnizam/internal/resources"
)

const (
	BulkOperationKind       = "BulkOperation"
	BulkOperationAPIVersion = "bulk.axiomnizam.io/v1"
)

// BulkOperationSpec is the desired state of a bulk operation.
type BulkOperationSpec struct {
	TenantID        string               `json:"tenantId"`
	Type            BulkOpType           `json:"type"`
	ResourceType    string               `json:"resourceType"`
	Items           []BulkItem           `json:"items"`
	Options         BulkOperationOptions `json:"options"`
	Timeout         int                  `json:"timeout,omitempty"`
	Atomic          bool                 `json:"atomic,omitempty"`
	RollbackOnError bool                 `json:"rollbackOnError,omitempty"`

	// Cancel, when true, asks the controller to cancel the operation.
	Cancel bool `json:"cancel,omitempty"`
	// RetryFailed, when true, asks the controller to retry failed items.
	RetryFailed bool `json:"retryFailed,omitempty"`
}

// BulkOperationResourceStatus extends the canonical object status
// with bulk-operation telemetry. Controller-owned.
type BulkOperationResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	OperationStatus BulkOpStatus      `json:"operationStatus"`
	TotalItems      int64             `json:"totalItems"`
	SuccessCount    int64             `json:"successCount"`
	FailureCount    int64             `json:"failureCount"`
	SkippedCount    int64             `json:"skippedCount"`
	Progress        int               `json:"progress"`
	StartedAt       *time.Time        `json:"startedAt,omitempty"`
	CompletedAt     *time.Time        `json:"completedAt,omitempty"`
	ErrorSummary    *BulkErrorSummary `json:"errorSummary,omitempty"`
}

// BulkOperationResource is the declarative resource for a BulkOperation.
type BulkOperationResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   BulkOperationSpec           `json:"spec"`
	Status BulkOperationResourceStatus `json:"status"`
}

// --- resources.Resource implementation ---

func (r *BulkOperationResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *BulkOperationResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *BulkOperationResource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *BulkOperationResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *BulkOperationResource) DeepCopy() resources.Resource {
	cp := *r
	if len(r.Spec.Items) > 0 {
		cp.Spec.Items = append([]BulkItem(nil), r.Spec.Items...)
	}
	return &cp
}

// --- reconciler.Resource implementation ---

func (r *BulkOperationResource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *BulkOperationResource) GetGeneration() int64         { return r.Generation }
func (r *BulkOperationResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }
