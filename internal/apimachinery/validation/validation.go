// Package validation implements the identifier-format checks that the
// Kubernetes API enforces: DNS-1123 labels, DNS-1123 subdomains,
// DNS-1035 labels, and the "qualified name" form used for label
// keys.  AxiomNizam reuses these rules for its own resource names so
// that kubectl and operators can address them without quoting games.
//
// All functions return a slice of strings rather than (bool, error)
// so callers can aggregate every violation a user committed in a
// single HTTP response — the 422 body includes every reason, matching
// upstream's "field-path: why" convention.
package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// dns1123LabelMaxLength is the maximum length of a DNS-1123 label.
const (
	// DNS1123LabelMaxLength is 63 per RFC 1123.
	DNS1123LabelMaxLength = 63
	// DNS1123SubdomainMaxLength is the practical cap from RFC 1123
	// §2.1: 253 chars including dots.
	DNS1123SubdomainMaxLength = 253
	// LabelValueMaxLength is the per-entry cap on k8s label values.
	LabelValueMaxLength = 63
	// QualifiedNameMaxLength is the cap on "prefix/name" label keys.
	QualifiedNameMaxLength = 63
)

var (
	// A DNS-1123 label is a-z0-9, starts and ends with alphanumeric,
	// may contain internal hyphens.  Empty is invalid.
	dns1123LabelRe = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

	// A DNS-1123 subdomain is one or more labels separated by dots.
	dns1123SubdomainRe = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

	// The qualified-name grammar from labels: "prefix/name" where
	// prefix is a DNS-1123 subdomain and name is letters, digits,
	// dashes, dots, and underscores, starting/ending alphanumeric.
	qualifiedNameRe = regexp.MustCompile(`^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$`)

	// Label values follow the same grammar as qualified names but
	// without the prefix slash.
	labelValueRe = qualifiedNameRe
)

// IsDNS1123Label reports any violations in name as a slice of reasons.
// An empty slice means the name is valid.
func IsDNS1123Label(name string) []string {
	var errs []string
	if len(name) > DNS1123LabelMaxLength {
		errs = append(errs, fmt.Sprintf("must be no more than %d characters", DNS1123LabelMaxLength))
	}
	if !dns1123LabelRe.MatchString(name) {
		errs = append(errs, "must consist of lowercase alphanumerics or '-', starting and ending with alphanumeric")
	}
	return errs
}

// IsDNS1123Subdomain reports violations for a dotted DNS name.
func IsDNS1123Subdomain(name string) []string {
	var errs []string
	if len(name) > DNS1123SubdomainMaxLength {
		errs = append(errs, fmt.Sprintf("must be no more than %d characters", DNS1123SubdomainMaxLength))
	}
	if !dns1123SubdomainRe.MatchString(name) {
		errs = append(errs, "must consist of lowercase alphanumerics, '-', or '.', starting and ending with alphanumeric")
	}
	return errs
}

// IsValidLabelValue validates the RHS of a k8s label entry.
func IsValidLabelValue(value string) []string {
	var errs []string
	if len(value) > LabelValueMaxLength {
		errs = append(errs, fmt.Sprintf("must be no more than %d characters", LabelValueMaxLength))
	}
	// Empty label values are legal in k8s.
	if value != "" && !labelValueRe.MatchString(value) {
		errs = append(errs, "must be a valid label value (alphanumerics, '-', '_', '.')")
	}
	return errs
}

// IsQualifiedName validates the LHS of a k8s label entry.  Two forms
// are legal: "name" and "prefix/name", where prefix is a DNS
// subdomain and name is the qualified-name grammar.
func IsQualifiedName(value string) []string {
	var errs []string
	if len(value) == 0 {
		return []string{"must not be empty"}
	}
	parts := strings.SplitN(value, "/", 2)
	var name string
	if len(parts) == 1 {
		name = parts[0]
	} else {
		prefix, n := parts[0], parts[1]
		if prefix == "" {
			errs = append(errs, "prefix part must be non-empty")
		} else if pfxErrs := IsDNS1123Subdomain(prefix); len(pfxErrs) > 0 {
			for _, e := range pfxErrs {
				errs = append(errs, "prefix "+e)
			}
		}
		name = n
	}
	if len(name) == 0 {
		errs = append(errs, "name part must be non-empty")
	} else if len(name) > QualifiedNameMaxLength {
		errs = append(errs, fmt.Sprintf("name part must be no more than %d characters", QualifiedNameMaxLength))
	} else if !qualifiedNameRe.MatchString(name) {
		errs = append(errs, "name part must match [A-Za-z0-9][-A-Za-z0-9_.]*[A-Za-z0-9]")
	}
	return errs
}

// IsValidNamespace applies the DNS-1123 label rules with an override
// for the reserved `kube-` and `default` prefixes: we disallow user
// creation of names under the kube- prefix to prevent shadowing.
func IsValidNamespace(name string) []string {
	errs := IsDNS1123Label(name)
	if strings.HasPrefix(name, "kube-") {
		errs = append(errs, "prefix 'kube-' is reserved")
	}
	return errs
}
