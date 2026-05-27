package rules

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// RuleListResponse is the response for listing quality rules.
type RuleListResponse struct {
	Rules []*QualityRuleResource `json:"rules"`
	Count int                    `json:"count"`
}

// RunRuleResponse is the response for a manual rule evaluation.
type RunRuleResponse struct {
	Rule   string       `json:"rule"`
	Result *CheckOutput `json:"result"`
}

// CheckListResponse is the response for listing check results.
type CheckListResponse struct {
	Checks []*QualityCheckResource `json:"checks"`
	Count  int                     `json:"count"`
}

// AssetScoreResponse is the response for an asset quality score.
type AssetScoreResponse struct {
	Asset         string  `json:"asset"`
	Score         float64 `json:"score"`
	TotalRules    int     `json:"totalRules"`
	PassingRules  int     `json:"passingRules"`
	CriticalFails int     `json:"criticalFails"`
	WarningFails  int     `json:"warningFails"`
}

// QualitySummaryResponse is the response for a platform-wide quality summary.
type QualitySummaryResponse struct {
	TotalRules   int     `json:"totalRules"`
	Passing      int     `json:"passing"`
	Failing      int     `json:"failing"`
	Errors       int     `json:"errors"`
	Pending      int     `json:"pending"`
	OverallScore float64 `json:"overallScore"`
}
