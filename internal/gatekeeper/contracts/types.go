package contracts

import "github.com/google/uuid"

// ChallengeResult contains the result of a challenge operation.
type ChallengeResult struct {
	ChallengeID uuid.UUID
	Verified    bool
	ExpiresAt   string
}

// DeviceTrustResult contains the result of device trust operation.
type DeviceTrustResult struct {
	DeviceID uuid.UUID
	Token    string
}

// RiskAssessment contains the result of a risk assessment.
type RiskAssessment struct {
	Score      int
	Level      string
	IsHighRisk bool
	Reason     string
}

// PolicyDecision contains the result of a policy evaluation.
type PolicyDecision struct {
	RequiresMFA    bool
	AllowedFactors []string
	RiskAction     string
	Reason         string
}
