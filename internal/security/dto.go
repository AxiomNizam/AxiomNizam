package security

import "time"

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// TLSCertStatusResponse is returned by the TLS certificate status endpoint.
type TLSCertStatusResponse struct {
	Mode                   string              `json:"mode"`
	Status                 string              `json:"status"`
	ThresholdDays          int                 `json:"threshold_days"`
	RenewCommandConfigured bool                `json:"renew_command_configured"`
	HasError               bool                `json:"has_error"`
	Certificates           []CertificateStatus `json:"certificates"`
}

// KubeadmCertStatusResponse is returned by the kubeadm certificate status endpoint.
type KubeadmCertStatusResponse struct {
	Mode                   string              `json:"mode"`
	Status                 string              `json:"status"`
	PKIDir                 string              `json:"pki_dir"`
	ManagedCerts           []string            `json:"managed_certs"`
	ThresholdDays          int                 `json:"threshold_days"`
	RenewCommandConfigured bool                `json:"renew_command_configured"`
	HasError               bool                `json:"has_error"`
	Certificates           []CertificateStatus `json:"certificates"`
}

// DryRunResponse is returned when a renewal command is prepared but not executed.
type DryRunResponse struct {
	Mode    string   `json:"mode,omitempty"`
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Target  string   `json:"target,omitempty"`
	Cert    string   `json:"cert,omitempty"`
	Command []string `json:"command"`
}

// RenewErrorResponse is returned when a renewal command fails.
type RenewErrorResponse struct {
	Mode    string   `json:"mode,omitempty"`
	Error   string   `json:"error"`
	Target  string   `json:"target,omitempty"`
	Cert    string   `json:"cert,omitempty"`
	Command []string `json:"command"`
	Output  string   `json:"output"`
}

// RenewSuccessResponse is returned after a successful certificate renewal command.
type RenewSuccessResponse struct {
	Mode              string              `json:"mode"`
	Status            string              `json:"status"`
	Message           string              `json:"message"`
	Target            string              `json:"target,omitempty"`
	Cert              string              `json:"cert,omitempty"`
	Command           []string            `json:"command"`
	Output            string              `json:"output"`
	CertificateStatus *CertificateStatus  `json:"certificate_status,omitempty"`
	Certificates      []CertificateStatus `json:"certificates,omitempty"`
}

// CheckRowAccessResponse is returned by the RLS row-access check endpoint.
type CheckRowAccessResponse struct {
	UserID    string    `json:"user_id"`
	Table     string    `json:"table"`
	Operation string    `json:"operation"`
	Allowed   bool      `json:"allowed"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// ListPoliciesResponse is returned by the RLS list-policies endpoint.
type ListPoliciesResponse struct {
	Table     string                   `json:"table"`
	Policies  []map[string]interface{} `json:"policies"`
	Count     int                      `json:"count"`
	Timestamp time.Time                `json:"timestamp"`
}

// SecurityStatsResponse is returned by the RLS security-stats endpoint.
type SecurityStatsResponse struct {
	TotalPolicies int       `json:"total_policies"`
	ActiveUsers   int       `json:"active_users"`
	AccessAllowed int       `json:"access_allowed"`
	AccessDenied  int       `json:"access_denied"`
	DenialRate    float64   `json:"denial_rate"`
	AuditLogSize  int       `json:"audit_log_size"`
	Timestamp     time.Time `json:"timestamp"`
}
