package rules

// =====================================================
// WS-2.1 — Quality Rules Alert Integration
//
// Bridges quality rule failures to the alerting system (WS-4).
// When a quality check fails and alertOnFailure is true, this
// dispatches an alert through the configured channels.
// =====================================================

import (
	"context"
	"fmt"
	"time"
)

// AlertSender abstracts sending alerts to the alerting system.
type AlertSender interface {
	// SendQualityAlert dispatches a quality failure alert.
	SendQualityAlert(ctx context.Context, alert QualityAlert) error
}

// QualityAlert represents an alert triggered by a quality rule failure.
type QualityAlert struct {
	RuleName       string            `json:"ruleName"`
	AssetRef       string            `json:"assetRef"`
	Severity       string            `json:"severity"`
	Message        string            `json:"message"`
	CheckResult    string            `json:"checkResult"`
	FailCount      int64             `json:"failCount,omitempty"`
	ConsecutiveFails int             `json:"consecutiveFails"`
	Channels       []string          `json:"channels"`
	FiredAt        time.Time         `json:"firedAt"`
	Labels         map[string]string `json:"labels,omitempty"`
}

// QualityAlertDispatcher dispatches quality alerts.
type QualityAlertDispatcher struct {
	sender AlertSender
}

// NewQualityAlertDispatcher creates a new dispatcher.
func NewQualityAlertDispatcher(sender AlertSender) *QualityAlertDispatcher {
	return &QualityAlertDispatcher{sender: sender}
}

// CheckAndAlert evaluates whether an alert should be sent for a rule failure.
func (d *QualityAlertDispatcher) CheckAndAlert(ctx context.Context, rule *QualityRuleResource, output *CheckOutput) error {
	if d.sender == nil {
		return nil
	}

	// Only alert if configured and check failed.
	if !rule.Spec.AlertOnFailure {
		return nil
	}
	if output.Passed {
		return nil
	}

	// Build alert.
	alert := QualityAlert{
		RuleName:         rule.Name,
		AssetRef:         rule.Spec.AssetRef,
		Severity:         rule.Spec.Severity,
		Message:          output.Message,
		CheckResult:      string(CheckResultFail),
		FailCount:        output.FailCount,
		ConsecutiveFails: rule.Status.ConsecutiveFails + 1,
		Channels:         rule.Spec.AlertChannels,
		FiredAt:          time.Now(),
		Labels: map[string]string{
			"rule":      rule.Name,
			"asset":     rule.Spec.AssetRef,
			"ruleType":  string(rule.Spec.RuleType),
			"severity":  rule.Spec.Severity,
		},
	}

	// Check SLA breach for escalation.
	if rule.Spec.SLA != nil {
		if rule.Status.ConsecutiveFails+1 >= rule.Spec.SLA.MaxConsecutiveFails {
			alert.Severity = "critical"
			alert.Message = fmt.Sprintf("[SLA BREACH] %s — %d consecutive failures (max: %d)",
				output.Message, rule.Status.ConsecutiveFails+1, rule.Spec.SLA.MaxConsecutiveFails)
			alert.Labels["sla_breach"] = "true"
		}
	}

	return d.sender.SendQualityAlert(ctx, alert)
}

// ShouldAlert determines if an alert should fire based on rule config and state.
func ShouldAlert(rule *QualityRuleResource, passed bool) bool {
	if !rule.Spec.AlertOnFailure {
		return false
	}
	if passed {
		return false
	}
	// Don't re-alert if already in failure state (dedup).
	// Only alert on transition from pass to fail, or on SLA breach.
	if rule.Status.LastResult == CheckResultFail && rule.Status.ConsecutiveFails > 1 {
		// Already alerted — only re-alert on SLA breach.
		if rule.Spec.SLA != nil && rule.Status.ConsecutiveFails >= rule.Spec.SLA.MaxConsecutiveFails {
			return true
		}
		return false
	}
	return true
}
