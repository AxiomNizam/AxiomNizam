package tenant

// =====================================================
// P2 resource-ification — Tenant.
//
// TenantV1Resource (named to avoid colliding with the legacy
// `TenantResource` in models.go, which is a per-tenant generic data
// container, not a tenant-lifecycle resource) is the declarative
// envelope around the imperative `Tenant` struct.  A dedicated
// reconciler drives tenant lifecycle (Active / Suspended / Archived /
// Deleting) through `TenantManager`.
// =====================================================

import (
	"context"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/reconciler"
	"example.com/axiomnizam/internal/resources"
)

const (
	TenantKind       = "Tenant"
	TenantAPIVersion = "tenant.axiomnizam.io/v1"
)

// TenantSpec is the desired state of a tenant.
type TenantSpec struct {
	DisplayName    string            `json:"displayName"`
	Description    string            `json:"description,omitempty"`
	Owner          string            `json:"owner"`
	Plan           string            `json:"plan,omitempty"`
	Tier           TenantTier        `json:"tier,omitempty"`
	IsolationLevel TenantIsolation   `json:"isolationLevel,omitempty"`
	DataLocation   string            `json:"dataLocation,omitempty"`
	Features       map[string]bool   `json:"features,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`

	// Suspended, when true, asks the controller to transition the
	// tenant to Suspended state.  Clearing it flips back to Active.
	Suspended bool `json:"suspended,omitempty"`
}

// TenantResourceStatus extends the canonical object status.
type TenantResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	// TenantStatus mirrors `Tenant.Status` (ACTIVE/SUSPENDED/…).
	TenantStatus TenantStatus `json:"tenantStatus,omitempty"`

	// SuspendedAt / DeletedAt are controller-recorded transitions.
	SuspendedAt *time.Time `json:"suspendedAt,omitempty"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

// TenantV1Resource is the declarative resource for a Tenant.
type TenantV1Resource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   TenantSpec           `json:"spec"`
	Status TenantResourceStatus `json:"status"`
}

// --- resources.Resource implementation ---

func (r *TenantV1Resource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *TenantV1Resource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *TenantV1Resource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *TenantV1Resource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *TenantV1Resource) DeepCopy() resources.Resource { cp := *r; return &cp }

// --- reconciler.Resource implementation ---

func (r *TenantV1Resource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *TenantV1Resource) GetGeneration() int64         { return r.Generation }
func (r *TenantV1Resource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }

// TenantReconciler reconciles TenantV1Resource objects.
//
// Skeleton only: it computes the target TenantStatus from Spec.Suspended
// and records it on the status.  A production reconciler also drives
// quota setup, DB provisioning, and soft-delete timing through
// TenantManager.
type TenantReconciler struct {
	store store.ResourceStore[*TenantV1Resource]
}

// NewTenantReconciler builds a reconciler.
func NewTenantReconciler(rs store.ResourceStore[*TenantV1Resource]) *TenantReconciler {
	return &TenantReconciler{store: rs}
}

// Reconcile implements `reconciler.Reconciler`.
func (r *TenantReconciler) Reconcile(ctx context.Context, obj reconciler.Resource) reconciler.ReconcileResult {
	res, ok := obj.(*TenantV1Resource)
	if !ok {
		return reconciler.ReconcileResult{Error: tenantErr("tenant: reconciler received non-TenantV1Resource")}
	}

	now := time.Now()
	status := res.Status
	target := TenantActive
	if res.Spec.Suspended {
		target = TenantSuspended
		if status.SuspendedAt == nil {
			status.SuspendedAt = &now
		}
	} else {
		status.SuspendedAt = nil
	}
	if res.DeletedAt != nil && status.DeletedAt == nil {
		status.DeletedAt = res.DeletedAt
		target = TenantDeleting
	}

	status.TenantStatus = target
	status.Phase = string(target)
	status.ObservedGeneration = res.Generation
	status.LastTransitionTime = now
	res.Status = status

	if r.store != nil {
		_ = r.store.Update(ctx, res)
	}
	return reconciler.ReconcileResult{}
}

type tenantErr string

func (e tenantErr) Error() string { return string(e) }
