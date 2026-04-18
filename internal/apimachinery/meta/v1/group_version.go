// Package metav1 — GroupVersion types.
//
// GroupVersionKind addresses a schema type; GroupVersionResource
// addresses a REST URL segment (plural kind).  Controllers and the
// discovery endpoint use the latter.
package metav1

import "strings"

// GroupVersion is the pair.  It is the natural key in a Scheme.
type GroupVersion struct {
	Group, Version string
}

// String renders "group/version" or just "version" for the core group.
func (gv GroupVersion) String() string {
	if gv.Group == "" {
		return gv.Version
	}
	return gv.Group + "/" + gv.Version
}

// WithKind pairs this group/version with a Kind.
func (gv GroupVersion) WithKind(kind string) GroupVersionKind {
	return GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: kind}
}

// WithResource pairs this group/version with a resource name (plural).
func (gv GroupVersion) WithResource(resource string) GroupVersionResource {
	return GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource}
}

// ParseGroupVersion parses "group/version" or "version".
func ParseGroupVersion(s string) GroupVersion {
	if i := strings.Index(s, "/"); i >= 0 {
		return GroupVersion{Group: s[:i], Version: s[i+1:]}
	}
	return GroupVersion{Version: s}
}

// GroupVersionKind is the triple used to identify Go types in a Scheme.
type GroupVersionKind struct {
	Group, Version, Kind string
}

// String renders "group/version, Kind=Foo".
func (gvk GroupVersionKind) String() string {
	return gvk.GroupVersion().String() + ", Kind=" + gvk.Kind
}

// GroupVersion returns the gv half.
func (gvk GroupVersionKind) GroupVersion() GroupVersion {
	return GroupVersion{Group: gvk.Group, Version: gvk.Version}
}

// GroupKind drops the version — useful for conversion lookups.
func (gvk GroupVersionKind) GroupKind() GroupKind {
	return GroupKind{Group: gvk.Group, Kind: gvk.Kind}
}

// Empty reports whether the triple is the zero value.
func (gvk GroupVersionKind) Empty() bool { return gvk == GroupVersionKind{} }

// GroupKind is the version-agnostic pair.
type GroupKind struct {
	Group, Kind string
}

// String renders "Kind.group" or just "Kind" for core group.
func (gk GroupKind) String() string {
	if gk.Group == "" {
		return gk.Kind
	}
	return gk.Kind + "." + gk.Group
}

// WithVersion pairs this GroupKind with a version.
func (gk GroupKind) WithVersion(version string) GroupVersionKind {
	return GroupVersionKind{Group: gk.Group, Version: version, Kind: gk.Kind}
}

// GroupVersionResource is the REST-addressable triple.  Resource is
// the plural lowercase form (e.g. "apibanks" rather than "APIBank").
type GroupVersionResource struct {
	Group, Version, Resource string
}

// String renders "group/version, Resource=foo".
func (gvr GroupVersionResource) String() string {
	return GroupVersion{Group: gvr.Group, Version: gvr.Version}.String() + ", Resource=" + gvr.Resource
}

// GroupResource drops the version.
func (gvr GroupVersionResource) GroupResource() GroupResource {
	return GroupResource{Group: gvr.Group, Resource: gvr.Resource}
}

// GroupResource is the version-agnostic REST pair.
type GroupResource struct {
	Group, Resource string
}

// String renders "resource.group" or just "resource" for core group.
func (gr GroupResource) String() string {
	if gr.Group == "" {
		return gr.Resource
	}
	return gr.Resource + "." + gr.Group
}
