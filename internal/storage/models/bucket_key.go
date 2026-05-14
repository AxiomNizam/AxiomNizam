package models

import "fmt"

// GetKey returns "tenantID/name" (or bare name when tenantID is empty).
// Satisfies the reconciler.Resource contract (P0.1).
func (b *BucketResource) GetKey() string {
	if b.Metadata.TenantID == "" {
		return b.Metadata.Name
	}
	return fmt.Sprintf("%s/%s", b.Metadata.TenantID, b.Metadata.Name)
}

// GetGeneration returns the spec generation. Satisfies reconciler.Resource.
func (b *BucketResource) GetGeneration() int64 {
	return b.Generation
}

// GetObservedGeneration returns the generation last observed by the
// controller. Satisfies reconciler.Resource.
func (b *BucketResource) GetObservedGeneration() int64 {
	return b.Status.ObservedGeneration
}
