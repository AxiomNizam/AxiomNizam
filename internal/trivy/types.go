package trivy

import "time"

type TargetKind string

const (
	TargetImage TargetKind = "image"
	TargetFS    TargetKind = "fs"
	TargetK8s   TargetKind = "k8s"
	TargetRepo  TargetKind = "repo"
)

type OutputFormat string

const (
	FormatJSON  OutputFormat = "json"
	FormatTable OutputFormat = "table"
)

const (
	SeverityUnknown  = "UNKNOWN"
	SeverityCritical = "CRITICAL"
	SeverityHigh     = "HIGH"
	SeverityMedium   = "MEDIUM"
	SeverityLow      = "LOW"
)

type PolicyHook func(finding Finding) bool

type ScanRequest struct {
	Kind          TargetKind
	Target        string
	Severity      []string
	IgnoreUnfixed bool
	Timeout       time.Duration
	Format        OutputFormat
	UseExternal   bool
	RetryCount    int
	RetryBackoff  time.Duration
	PolicyHooks   []PolicyHook
}

type ScanResult struct {
	Scanner      string            `json:"scanner"`
	TargetKind   string            `json:"targetKind"`
	Target       string            `json:"target"`
	ArtifactName string            `json:"artifactName,omitempty"`
	ArtifactType string            `json:"artifactType,omitempty"`
	ScannedAt    time.Time         `json:"scannedAt"`
	Findings     []Finding         `json:"findings"`
	Summary      Summary           `json:"summary"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type Finding struct {
	Category         string `json:"category"`
	ID               string `json:"id"`
	Severity         string `json:"severity"`
	Title            string `json:"title,omitempty"`
	Description      string `json:"description,omitempty"`
	Target           string `json:"target,omitempty"`
	Resource         string `json:"resource,omitempty"`
	PackageName      string `json:"packageName,omitempty"`
	InstalledVersion string `json:"installedVersion,omitempty"`
	FixedVersion     string `json:"fixedVersion,omitempty"`
	Reference        string `json:"reference,omitempty"`
	Status           string `json:"status,omitempty"`
	Unfixed          bool   `json:"unfixed"`
}

type Summary struct {
	Total    int `json:"total"`
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Unknown  int `json:"unknown"`
}
