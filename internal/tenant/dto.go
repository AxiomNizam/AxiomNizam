package tenant

import "time"

// CreateTenantRequest is the API request for creating a tenant.
type CreateTenantRequest struct {
	Name           string          `json:"name" binding:"required"`
	DisplayName    string          `json:"displayName"`
	Owner          string          `json:"owner" binding:"required"`
	Tier           TenantTier      `json:"tier"`
	IsolationLevel TenantIsolation `json:"isolationLevel"`
}

// UpdateTenantRequest is the API request for updating a tenant.
type UpdateTenantRequest struct {
	DisplayName    *string          `json:"displayName,omitempty"`
	Description    *string          `json:"description,omitempty"`
	Tier           *TenantTier      `json:"tier,omitempty"`
	IsolationLevel *TenantIsolation `json:"isolationLevel,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Features       map[string]bool   `json:"features,omitempty"`
}

// TenantResponse is the API response for a tenant.
type TenantResponse struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	DisplayName    string            `json:"displayName,omitempty"`
	Owner          string            `json:"owner,omitempty"`
	Tier           TenantTier        `json:"tier"`
	IsolationLevel TenantIsolation   `json:"isolationLevel,omitempty"`
	Status         TenantStatus      `json:"status"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// TenantListResponse is the API response for listing tenants.
type TenantListResponse struct {
	Tenants []*Tenant `json:"tenants"`
	Count   int       `json:"count"`
}

// TenantCreatedResponse is the accepted response for reconciler-authoritative creation.
type TenantCreatedResponse struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// AddMemberRequest is the API request for adding a tenant member.
type AddMemberRequest struct {
	UserID string     `json:"userId" binding:"required"`
	Role   MemberRole `json:"role" binding:"required"`
}

// CheckQuotaRequest is the API request for checking quota.
type CheckQuotaRequest struct {
	Resource string `json:"resource" binding:"required"`
	Amount   int64  `json:"amount" binding:"required"`
}

// QuotaCheckResponse is the API response for quota check.
type QuotaCheckResponse struct {
	Allowed bool `json:"allowed"`
}

// MessageResponse is a generic action acknowledgment.
type MessageResponse struct {
	Message string `json:"message"`
}
