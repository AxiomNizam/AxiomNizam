package risk

import (
	"net"
	"time"
)

// SignalBuilder provides a fluent API for constructing Signals.
type SignalBuilder struct {
	signals Signals
}

// NewSignalBuilder creates a new signal builder.
func NewSignalBuilder() *SignalBuilder {
	return &SignalBuilder{}
}

// WithIP sets the IP address and detects if it's a datacenter IP.
func (b *SignalBuilder) WithIP(ip string) *SignalBuilder {
	b.signals.IPAddress = ip
	b.signals.DatacenterIP = IsDatacenterIP(ip)
	return b
}

// WithDevice sets device-related signals.
func (b *SignalBuilder) WithDevice(isNew bool, fingerprint string) *SignalBuilder {
	b.signals.IsNewDevice = isNew
	b.signals.DeviceFingerprint = fingerprint
	return b
}

// WithBrowser sets browser-related signals.
func (b *SignalBuilder) WithBrowser(newBrowser bool) *SignalBuilder {
	b.signals.NewBrowser = newBrowser
	return b
}

// WithGeo sets geographic signals.
func (b *SignalBuilder) WithGeo(location string, diffKm int) *SignalBuilder {
	b.signals.GeoLocation = location
	b.signals.GeoDifference = diffKm
	return b
}

// WithBehavior sets behavioral signals.
func (b *SignalBuilder) WithBehavior(unusual, suspicious bool) *SignalBuilder {
	b.signals.UnusualActivity = unusual
	b.signals.SuspiciousLogin = suspicious
	return b
}

// WithFailures sets failure tracking signals.
func (b *SignalBuilder) WithFailures(count int, window time.Duration) *SignalBuilder {
	b.signals.FrequentFailures = count > 0
	b.signals.FailureCount = count
	b.signals.FailureWindow = window
	return b
}

// WithAccountAge sets account maturity signals.
func (b *SignalBuilder) WithAccountAge(age time.Duration) *SignalBuilder {
	b.signals.AccountAge = age
	return b
}

// WithVPN sets VPN detection signal.
func (b *SignalBuilder) WithVPN(detected bool) *SignalBuilder {
	b.signals.VPNDetected = detected
	return b
}

// WithIPReputation sets the IP reputation score.
func (b *SignalBuilder) WithIPReputation(score int) *SignalBuilder {
	b.signals.IPReputation = score
	return b
}

// WithPrivilege sets privilege-related signals.
func (b *SignalBuilder) WithPrivilege(highPriv, sensitive bool) *SignalBuilder {
	b.signals.HighPrivilegeOp = highPriv
	b.signals.SensitiveAction = sensitive
	return b
}

// Build returns the constructed Signals.
func (b *SignalBuilder) Build() *Signals {
	return &b.signals
}

// IsDatacenterIP checks if an IP belongs to known datacenter ranges.
func IsDatacenterIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	// Common cloud provider ranges (simplified)
	datacenterPrefixes := []string{
		"3.", "34.", "35.", "52.", "54.", "13.", "18.", "52.",
		"104.", "130.", "131.", "135.", "137.", "138.",
	}
	for _, prefix := range datacenterPrefixes {
		if len(ip) >= len(prefix) && ip[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// DaysSince calculates days between two times.
func DaysSince(t time.Time) int {
	return int(time.Since(t).Hours() / 24)
}
