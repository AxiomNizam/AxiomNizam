package resources

import "fmt"

// GetKey returns the canonical key for the resource in the form "namespace/name".
// When Namespace is empty the bare name is returned.
//
// This method makes any type that embeds *BaseResource (or BaseResource) auto-
// satisfy the reconciler.Resource interface without requiring each concrete
// resource type to implement these accessors manually.
func (br *BaseResource) GetKey() string {
	if br.Namespace == "" {
		return br.Name
	}
	return fmt.Sprintf("%s/%s", br.Namespace, br.Name)
}

// GetGeneration returns the spec generation of the resource.
// Part of the reconciler.Resource contract.
func (br *BaseResource) GetGeneration() int64 {
	return br.ObjectMeta.Generation
}

// GetObservedGeneration returns the generation last observed by the controller.
// Part of the reconciler.Resource contract.
func (br *BaseResource) GetObservedGeneration() int64 {
	return br.Status.ObservedGeneration
}
