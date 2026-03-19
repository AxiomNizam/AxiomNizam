package waitx

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	DNSRecordA     = "A"
	DNSRecordAAAA  = "AAAA"
	DNSRecordCNAME = "CNAME"
	DNSRecordMX    = "MX"
	DNSRecordTXT   = "TXT"
	DNSRecordNS    = "NS"
)

// DNSChecker verifies DNS records for a domain.
type DNSChecker struct {
	RecordType     string
	Address        string
	NameServer     string
	ExpectedValues []string
	DialTimeout    time.Duration
}

func (c DNSChecker) Name() string {
	recordType := strings.ToUpper(strings.TrimSpace(c.RecordType))
	return fmt.Sprintf("dns-%s:%s", recordType, strings.TrimSpace(c.Address))
}

func (c DNSChecker) Check(ctx context.Context) error {
	recordType := strings.ToUpper(strings.TrimSpace(c.RecordType))
	address := strings.TrimSpace(c.Address)
	if address == "" {
		return fmt.Errorf("dns address is required")
	}

	if !isSupportedDNSRecordType(recordType) {
		return fmt.Errorf("unsupported dns record type %q", c.RecordType)
	}

	resolver := c.resolver()
	values, err := lookupDNSValues(ctx, resolver, recordType, address)
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return fmt.Errorf("no %s records found for %s", recordType, address)
	}

	expected := sanitizeStringSlice(c.ExpectedValues)
	if len(expected) == 0 {
		return nil
	}

	normalize := dnsValueNormalizer(recordType)
	if !containsAll(values, expected, normalize) {
		return fmt.Errorf("dns %s records for %s did not include expected values (actual=%v expected=%v)", recordType, address, values, expected)
	}

	return nil
}

func (c DNSChecker) resolver() *net.Resolver {
	nameserver := strings.TrimSpace(c.NameServer)
	if nameserver == "" {
		return net.DefaultResolver
	}

	timeout := c.DialTimeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	nameserver = ensurePort(nameserver, "53")
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			dialer := net.Dialer{Timeout: timeout}
			return dialer.DialContext(ctx, network, nameserver)
		},
	}
}

func lookupDNSValues(ctx context.Context, resolver *net.Resolver, recordType, address string) ([]string, error) {
	switch recordType {
	case DNSRecordA:
		return lookupARecords(ctx, resolver, address)
	case DNSRecordAAAA:
		return lookupAAAARecords(ctx, resolver, address)
	case DNSRecordCNAME:
		return lookupCNAMERecords(ctx, resolver, address)
	case DNSRecordMX:
		return lookupMXRecords(ctx, resolver, address)
	case DNSRecordTXT:
		return lookupTXTRecords(ctx, resolver, address)
	case DNSRecordNS:
		return lookupNSRecords(ctx, resolver, address)
	default:
		return nil, fmt.Errorf("unsupported dns record type %q", recordType)
	}
}

func lookupARecords(ctx context.Context, resolver *net.Resolver, address string) ([]string, error) {
	ips, err := resolver.LookupIP(ctx, "ip4", address)
	if err != nil {
		return nil, fmt.Errorf("dns A lookup failed: %w", err)
	}
	return ipValues(ips), nil
}

func lookupAAAARecords(ctx context.Context, resolver *net.Resolver, address string) ([]string, error) {
	ips, err := resolver.LookupIP(ctx, "ip6", address)
	if err != nil {
		return nil, fmt.Errorf("dns AAAA lookup failed: %w", err)
	}
	return ipValues(ips), nil
}

func lookupCNAMERecords(ctx context.Context, resolver *net.Resolver, address string) ([]string, error) {
	value, err := resolver.LookupCNAME(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("dns CNAME lookup failed: %w", err)
	}
	return []string{strings.TrimSpace(value)}, nil
}

func lookupMXRecords(ctx context.Context, resolver *net.Resolver, address string) ([]string, error) {
	mxRecords, err := resolver.LookupMX(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("dns MX lookup failed: %w", err)
	}
	values := make([]string, 0, len(mxRecords))
	for _, mx := range mxRecords {
		if mx == nil {
			continue
		}
		values = append(values, strings.TrimSpace(mx.Host))
	}
	return values, nil
}

func lookupTXTRecords(ctx context.Context, resolver *net.Resolver, address string) ([]string, error) {
	txtValues, err := resolver.LookupTXT(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("dns TXT lookup failed: %w", err)
	}
	return sanitizeStringSlice(txtValues), nil
}

func lookupNSRecords(ctx context.Context, resolver *net.Resolver, address string) ([]string, error) {
	nsRecords, err := resolver.LookupNS(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("dns NS lookup failed: %w", err)
	}
	values := make([]string, 0, len(nsRecords))
	for _, ns := range nsRecords {
		if ns == nil {
			continue
		}
		values = append(values, strings.TrimSpace(ns.Host))
	}
	return values, nil
}

func isSupportedDNSRecordType(recordType string) bool {
	switch recordType {
	case DNSRecordA, DNSRecordAAAA, DNSRecordCNAME, DNSRecordMX, DNSRecordTXT, DNSRecordNS:
		return true
	default:
		return false
	}
}

func ipValues(ips []net.IP) []string {
	values := make([]string, 0, len(ips))
	for _, ip := range ips {
		if ip == nil {
			continue
		}
		values = append(values, ip.String())
	}
	return values
}

func sanitizeStringSlice(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

func dnsValueNormalizer(recordType string) func(string) string {
	recordType = strings.ToUpper(strings.TrimSpace(recordType))
	switch recordType {
	case DNSRecordCNAME, DNSRecordMX, DNSRecordNS:
		return normalizeDNSDomain
	case DNSRecordA, DNSRecordAAAA:
		return normalizeIP
	default:
		return func(value string) string {
			return strings.TrimSpace(value)
		}
	}
}

func normalizeDNSDomain(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimSuffix(trimmed, ".")
	return strings.ToLower(trimmed)
}

func normalizeIP(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return strings.TrimSpace(value)
	}
	return ip.String()
}

func containsAll(actual, expected []string, normalize func(string) string) bool {
	if normalize == nil {
		normalize = func(value string) string { return value }
	}

	actualSet := make(map[string]struct{}, len(actual))
	for _, value := range actual {
		normalized := normalize(value)
		if normalized == "" {
			continue
		}
		actualSet[normalized] = struct{}{}
	}

	for _, value := range expected {
		normalized := normalize(value)
		if normalized == "" {
			continue
		}
		if _, ok := actualSet[normalized]; !ok {
			return false
		}
	}

	return true
}

func ensurePort(address, defaultPort string) string {
	trimmed := strings.TrimSpace(address)
	if trimmed == "" {
		return ""
	}
	if _, _, err := net.SplitHostPort(trimmed); err == nil {
		return trimmed
	}

	if ip := net.ParseIP(trimmed); ip != nil {
		return net.JoinHostPort(trimmed, defaultPort)
	}

	if strings.Contains(trimmed, ":") && !strings.Contains(trimmed, "]") {
		return trimmed
	}

	return net.JoinHostPort(trimmed, defaultPort)
}
