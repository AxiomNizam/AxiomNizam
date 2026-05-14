// Package metav1 is AxiomNizam's thin stand-in for the upstream
// k8s.io/apimachinery/pkg/apis/meta/v1 package.  It defines the
// shared envelope fields every resource carries — TypeMeta (apiVersion
// + kind), ObjectMeta (identity + system fields), and ListMeta (list
// pagination continuation / revision).  The struct layouts are
// source-compatible with upstream JSON so that off-the-shelf kubectl
// tooling can consume AxiomNizam resources when they follow k8s
// conventions.
package metav1

// TypeMeta is inlined into every top-level object.  It lets a generic
// decoder dispatch a payload to the correct Go type without an
// out-of-band schema.
type TypeMeta struct {
	// Kind is the resource kind (e.g. "APIBank", "DataSource").
	Kind string `json:"kind,omitempty"`

	// APIVersion is "<group>/<version>" (e.g. "axiom.nizam/v1") or
	// just "<version>" for core group.
	APIVersion string `json:"apiVersion,omitempty"`
}

// GetObjectKind returns t — satisfies a minimal Object-kind accessor.
func (t *TypeMeta) GetObjectKind() *TypeMeta { return t }

// SetGroupVersionKind sets APIVersion + Kind from a GVK triple.
func (t *TypeMeta) SetGroupVersionKind(gvk GroupVersionKind) {
	t.Kind = gvk.Kind
	if gvk.Group == "" {
		t.APIVersion = gvk.Version
	} else {
		t.APIVersion = gvk.Group + "/" + gvk.Version
	}
}

// GroupVersionKind extracts the (group, version, kind) triple from
// the envelope.  The parse is permissive: malformed APIVersion yields
// empty Group and Version.
func (t TypeMeta) GroupVersionKind() GroupVersionKind {
	gvk := GroupVersionKind{Kind: t.Kind}
	slash := -1
	for i := 0; i < len(t.APIVersion); i++ {
		if t.APIVersion[i] == '/' {
			slash = i
			break
		}
	}
	if slash < 0 {
		gvk.Version = t.APIVersion
	} else {
		gvk.Group = t.APIVersion[:slash]
		gvk.Version = t.APIVersion[slash+1:]
	}
	return gvk
}

// ObjectMeta is the per-object metadata envelope.  Not every field
// is used by every resource — AxiomNizam picks the subset it needs.
type ObjectMeta struct {
	Name              string            `json:"name,omitempty"`
	Namespace         string            `json:"namespace,omitempty"`
	UID               string            `json:"uid,omitempty"`
	ResourceVersion   string            `json:"resourceVersion,omitempty"`
	Generation        int64             `json:"generation,omitempty"`
	CreationTimestamp Time              `json:"creationTimestamp,omitempty"`
	DeletionTimestamp *Time             `json:"deletionTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	OwnerReferences   []OwnerReference  `json:"ownerReferences,omitempty"`
	Finalizers        []string          `json:"finalizers,omitempty"`
}

// OwnerReference is re-declared here (alongside the existing
// internal/apimachinery/owner package) so metav1 stays self-contained
// and the JSON tag set matches upstream exactly.
type OwnerReference struct {
	APIVersion         string `json:"apiVersion"`
	Kind               string `json:"kind"`
	Name               string `json:"name"`
	UID                string `json:"uid"`
	Controller         *bool  `json:"controller,omitempty"`
	BlockOwnerDeletion *bool  `json:"blockOwnerDeletion,omitempty"`
}

// ListMeta is the pagination envelope for collection responses.  The
// continuation token is opaque to clients — they must only pass it
// back verbatim as the `continue=` query parameter on the next page.
type ListMeta struct {
	ResourceVersion    string `json:"resourceVersion,omitempty"`
	Continue           string `json:"continue,omitempty"`
	RemainingItemCount *int64 `json:"remainingItemCount,omitempty"`
}
