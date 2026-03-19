package apiscanner

import (
	"fmt"
	"sort"
	"strings"
)

const discoveryCheckFingerprintHeaders = "fingerprint.headers"

type SupportedDiscoveryScanIDs struct {
	API    []string
	Domain []string
}

func GetSupportedDiscoveryScanIDs() SupportedDiscoveryScanIDs {
	return SupportedDiscoveryScanIDs{
		API:    sortedDiscoveryIDSet(supportedAPIDiscoveryCheckIDs()),
		Domain: sortedDiscoveryIDSet(supportedDomainDiscoveryCheckIDs()),
	}
}

type discoveryCheckFilter struct {
	include map[string]struct{}
	exclude map[string]struct{}
}

func newDiscoveryCheckFilter(includeIDs []string, excludeIDs []string) discoveryCheckFilter {
	return discoveryCheckFilter{
		include: buildDiscoveryIDSet(includeIDs),
		exclude: buildDiscoveryIDSet(excludeIDs),
	}
}

func (f discoveryCheckFilter) Allows(id string) bool {
	normalized := normalizeDiscoveryID(id)
	if normalized == "" {
		return false
	}

	if _, denied := f.exclude[normalized]; denied {
		return false
	}

	if len(f.include) == 0 {
		return true
	}

	_, allowed := f.include[normalized]
	return allowed
}

func normalizeDiscoveryID(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeDiscoveryIDs(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(values))
	for _, raw := range values {
		for _, part := range strings.Split(raw, ",") {
			id := normalizeDiscoveryID(part)
			if id == "" {
				continue
			}
			normalized = append(normalized, id)
		}
	}

	if len(normalized) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(normalized))
	unique := make([]string, 0, len(normalized))
	for _, id := range normalized {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	sort.Strings(unique)
	return unique
}

func validateDiscoveryIDSelection(includeIDs []string, excludeIDs []string, supported map[string]struct{}) error {
	for _, id := range includeIDs {
		if _, ok := supported[id]; !ok {
			return fmt.Errorf("unknown include scan ID %q (supported: %s)", id, strings.Join(sortedDiscoveryIDSet(supported), ", "))
		}
	}

	for _, id := range excludeIDs {
		if _, ok := supported[id]; !ok {
			return fmt.Errorf("unknown exclude scan ID %q (supported: %s)", id, strings.Join(sortedDiscoveryIDSet(supported), ", "))
		}
	}

	if len(includeIDs) == 0 || len(excludeIDs) == 0 {
		return nil
	}

	includeSet := buildDiscoveryIDSet(includeIDs)
	for _, id := range excludeIDs {
		if _, overlap := includeSet[id]; overlap {
			return fmt.Errorf("scan ID %q cannot be in both include and exclude lists", id)
		}
	}

	return nil
}

func buildDiscoveryIDSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, raw := range values {
		id := normalizeDiscoveryID(raw)
		if id == "" {
			continue
		}
		set[id] = struct{}{}
	}
	return set
}

func sortedDiscoveryIDSet(values map[string]struct{}) []string {
	items := make([]string, 0, len(values))
	for value := range values {
		items = append(items, value)
	}
	sort.Strings(items)
	return items
}
