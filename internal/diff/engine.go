package diff

import (
	"context"
	"fmt"
)

// DiffType represents the type of difference
type DiffType string

const (
	DiffTypeAdded    DiffType = "added"
	DiffTypeRemoved  DiffType = "removed"
	DiffTypeModified DiffType = "modified"
	DiffTypeNone     DiffType = "none"
)

// ResourceDiff represents differences between two resource states
type ResourceDiff struct {
	Kind               string       `json:"kind"`
	Name               string       `json:"name"`
	Changes            []*FieldDiff `json:"changes"`
	PoliciesAffected   []string     `json:"policiesAffected,omitempty"`
	WorkflowsTriggered []string     `json:"workflowsTriggered,omitempty"`
	Summary            string       `json:"summary"`
}

// FieldDiff represents a difference in a single field
type FieldDiff struct {
	Path     string      `json:"path"`
	Type     DiffType    `json:"type"`
	OldValue interface{} `json:"oldValue,omitempty"`
	NewValue interface{} `json:"newValue,omitempty"`
	Reason   string      `json:"reason,omitempty"`
}

// DiffEngine computes differences between resources
type DiffEngine struct {
	policyEvaluator PolicyEvaluator
	workflowMatcher WorkflowMatcher
}

// PolicyEvaluator checks which policies are affected
type PolicyEvaluator interface {
	AffectedBy(old, new map[string]interface{}) []string
	Blocks(policy string, new map[string]interface{}) (bool, string)
}

// WorkflowMatcher checks which workflows are triggered
type WorkflowMatcher interface {
	MatchingWorkflows(diff *ResourceDiff) []string
}

// NewDiffEngine creates a new diff engine
func NewDiffEngine(pe PolicyEvaluator, wm WorkflowMatcher) *DiffEngine {
	return &DiffEngine{
		policyEvaluator: pe,
		workflowMatcher: wm,
	}
}

// Compute computes the diff between old and new resource states
func (de *DiffEngine) Compute(ctx context.Context, kind, name string, old, new map[string]interface{}) (*ResourceDiff, error) {
	diff := &ResourceDiff{
		Kind:    kind,
		Name:    name,
		Changes: make([]*FieldDiff, 0),
	}

	// Handle creation (no old state)
	if old == nil || len(old) == 0 {
		diff.Summary = fmt.Sprintf("Creating new %s: %s", kind, name)
		diff.Changes = append(diff.Changes, &FieldDiff{
			Path:     ".",
			Type:     DiffTypeAdded,
			NewValue: new,
			Reason:   "new resource",
		})
		return diff, nil
	}

	// Handle deletion (no new state)
	if new == nil || len(new) == 0 {
		diff.Summary = fmt.Sprintf("Deleting %s: %s", kind, name)
		diff.Changes = append(diff.Changes, &FieldDiff{
			Path:     ".",
			Type:     DiffTypeRemoved,
			OldValue: old,
			Reason:   "resource deleted",
		})
		return diff, nil
	}

	// Compare spec fields
	oldSpec := getNestedMap(old, "spec")
	newSpec := getNestedMap(new, "spec")

	fieldDiffs := computeMapDiff(oldSpec, newSpec, "spec")
	diff.Changes = append(diff.Changes, fieldDiffs...)

	// Check which policies are affected
	if de.policyEvaluator != nil {
		affected := de.policyEvaluator.AffectedBy(old, new)
		diff.PoliciesAffected = affected

		// Check if any policies block this change
		for _, policy := range affected {
			if blocks, reason := de.policyEvaluator.Blocks(policy, new); blocks {
				diff.Changes = append(diff.Changes, &FieldDiff{
					Path:   "policy",
					Type:   DiffTypeModified,
					Reason: fmt.Sprintf("Policy %s blocks: %s", policy, reason),
				})
			}
		}
	}

	// Check which workflows are triggered
	if de.workflowMatcher != nil {
		workflows := de.workflowMatcher.MatchingWorkflows(diff)
		diff.WorkflowsTriggered = workflows
	}

	// Generate summary
	diff.Summary = summarizeDiff(diff)

	return diff, nil
}

// computeMapDiff computes differences between two maps
func computeMapDiff(old, new map[string]interface{}, prefix string) []*FieldDiff {
	diffs := make([]*FieldDiff, 0)

	// Check for additions and modifications
	for key, newVal := range new {
		path := fmt.Sprintf("%s.%s", prefix, key)
		if oldVal, exists := old[key]; !exists {
			diffs = append(diffs, &FieldDiff{
				Path:     path,
				Type:     DiffTypeAdded,
				NewValue: newVal,
				Reason:   "new field",
			})
		} else if oldVal != newVal {
			diffs = append(diffs, &FieldDiff{
				Path:     path,
				Type:     DiffTypeModified,
				OldValue: oldVal,
				NewValue: newVal,
				Reason:   "field modified",
			})
		}
	}

	// Check for removals
	for key, oldVal := range old {
		if _, exists := new[key]; !exists {
			path := fmt.Sprintf("%s.%s", prefix, key)
			diffs = append(diffs, &FieldDiff{
				Path:     path,
				Type:     DiffTypeRemoved,
				OldValue: oldVal,
				Reason:   "field removed",
			})
		}
	}

	return diffs
}

// summarizeDiff creates a human-readable summary of changes
func summarizeDiff(diff *ResourceDiff) string {
	if len(diff.Changes) == 0 {
		return fmt.Sprintf("No changes to %s: %s", diff.Kind, diff.Name)
	}

	added := 0
	removed := 0
	modified := 0

	for _, change := range diff.Changes {
		switch change.Type {
		case DiffTypeAdded:
			added++
		case DiffTypeRemoved:
			removed++
		case DiffTypeModified:
			modified++
		}
	}

	summary := fmt.Sprintf("Changes to %s: %s\n", diff.Kind, diff.Name)
	if added > 0 {
		summary += fmt.Sprintf("  Added: %d fields\n", added)
	}
	if modified > 0 {
		summary += fmt.Sprintf("  Modified: %d fields\n", modified)
	}
	if removed > 0 {
		summary += fmt.Sprintf("  Removed: %d fields\n", removed)
	}

	if len(diff.PoliciesAffected) > 0 {
		summary += fmt.Sprintf("  Policies affected: %v\n", diff.PoliciesAffected)
	}

	if len(diff.WorkflowsTriggered) > 0 {
		summary += fmt.Sprintf("  Workflows triggered: %v\n", diff.WorkflowsTriggered)
	}

	return summary
}

// getNestedMap safely gets a nested map from a map
func getNestedMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if nm, ok := v.(map[string]interface{}); ok {
			return nm
		}
	}
	return make(map[string]interface{})
}

// Diff computes a basic diff without policy/workflow info.
// For full diff capabilities, create a DiffEngine instance directly.
func Diff(ctx context.Context, kind, name string, old, new map[string]interface{}) (*ResourceDiff, error) {
	engine := NewDiffEngine(nil, nil)
	return engine.Compute(ctx, kind, name, old, new)
}

// PrintDiff prints a diff in human-readable format
func PrintDiff(diff *ResourceDiff) string {
	output := fmt.Sprintf("Resource: %s/%s\n", diff.Kind, diff.Name)
	output += fmt.Sprintf("Summary: %s\n", diff.Summary)

	if len(diff.Changes) > 0 {
		output += "\nChanges:\n"
		for i, change := range diff.Changes {
			output += fmt.Sprintf("  %d. %s (%s)\n", i+1, change.Path, change.Type)
			if change.OldValue != nil {
				output += fmt.Sprintf("     Old: %v\n", change.OldValue)
			}
			if change.NewValue != nil {
				output += fmt.Sprintf("     New: %v\n", change.NewValue)
			}
			if change.Reason != "" {
				output += fmt.Sprintf("     Reason: %s\n", change.Reason)
			}
		}
	}

	if len(diff.PoliciesAffected) > 0 {
		output += fmt.Sprintf("\nAffected Policies: %v\n", diff.PoliciesAffected)
	}

	if len(diff.WorkflowsTriggered) > 0 {
		output += fmt.Sprintf("Triggered Workflows: %v\n", diff.WorkflowsTriggered)
	}

	return output
}
