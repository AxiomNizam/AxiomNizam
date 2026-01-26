package tenant

import "time"

// Tenant represents organization/workspace
type Tenant struct {
	ID             string            `json:"id" db:"id"`
	Name           string            `json:"name" db:"name"`
	DisplayName    string            `json:"displayName" db:"display_name"`
	Description    string            `json:"description" db:"description"`
	Owner          string            `json:"owner" db:"owner"`   // User ID
	Status         TenantStatus      `json:"status" db:"status"` // Active/Suspended/Archived
	Plan           string            `json:"plan" db:"plan"`     // Free/Pro/Enterprise
	Tier           TenantTier        `json:"tier" db:"tier"`     // Limits tier
	CreatedAt      time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time         `json:"updatedAt" db:"updated_at"`
	SuspendedAt    *time.Time        `json:"suspendedAt" db:"suspended_at"`
	DeletedAt      *time.Time        `json:"deletedAt" db:"deleted_at"`           // Soft delete
	Metadata       map[string]string `json:"metadata" db:"-"`                     // Custom metadata
	Features       map[string]bool   `json:"features" db:"-"`                     // Feature flags
	IsolationLevel TenantIsolation   `json:"isolationLevel" db:"isolation_level"` // Data isolation strategy
	DataLocation   string            `json:"dataLocation" db:"data_location"`     // Geographic region
}

// TenantStatus represents tenant lifecycle status
type TenantStatus string

const (
	TenantActive    TenantStatus = "ACTIVE"
	TenantSuspended TenantStatus = "SUSPENDED"
	TenantArchived  TenantStatus = "ARCHIVED"
	TenantDeleting  TenantStatus = "DELETING"
)

// TenantTier defines resource limits
type TenantTier string

const (
	TierFree       TenantTier = "FREE"
	TierPro        TenantTier = "PRO"
	TierEnterprise TenantTier = "ENTERPRISE"
)

// TenantIsolation defines isolation level
type TenantIsolation string

const (
	IsolationLogical  TenantIsolation = "LOGICAL"  // Shared database, filtered by tenant_id
	IsolationDatabase TenantIsolation = "DATABASE" // Separate database per tenant
	IsolationHybrid   TenantIsolation = "HYBRID"   // Critical data isolated, other shared
)

// TenantQuota represents resource limits
type TenantQuota struct {
	TenantID      string    `json:"tenantId" db:"tenant_id"`
	MaxUsers      int       `json:"maxUsers" db:"max_users"`
	MaxResources  int       `json:"maxResources" db:"max_resources"`
	MaxQueries    int64     `json:"maxQueries" db:"max_queries"`       // Per month
	MaxStorage    int64     `json:"maxStorage" db:"max_storage"`       // In bytes
	MaxAPIcalls   int64     `json:"maxApiCalls" db:"max_api_calls"`    // Per day
	MaxConcurrent int       `json:"maxConcurrent" db:"max_concurrent"` // Concurrent requests
	QueryTimeout  int       `json:"queryTimeout" db:"query_timeout"`   // Seconds
	UsedUsers     int       `json:"usedUsers" db:"used_users"`
	UsedResources int       `json:"usedResources" db:"used_resources"`
	UsedQueries   int64     `json:"usedQueries" db:"used_queries"`
	UsedStorage   int64     `json:"usedStorage" db:"used_storage"`
	UsedAPICalls  int64     `json:"usedApiCalls" db:"used_api_calls"`
	ResetDate     time.Time `json:"resetDate" db:"reset_date"` // When quota resets
}

// TenantMember represents user membership in tenant
type TenantMember struct {
	ID          string       `json:"id" db:"id"`
	TenantID    string       `json:"tenantId" db:"tenant_id"`
	UserID      string       `json:"userId" db:"user_id"`
	Username    string       `json:"username" db:"username"`
	Email       string       `json:"email" db:"email"`
	Role        MemberRole   `json:"role" db:"role"`     // Owner/Admin/Member/Viewer
	Status      MemberStatus `json:"status" db:"status"` // Active/Invited/Suspended
	JoinedAt    time.Time    `json:"joinedAt" db:"joined_at"`
	InvitedAt   *time.Time   `json:"invitedAt" db:"invited_at"`
	Permissions []string     `json:"permissions" db:"-"` // Custom permissions
}

// MemberRole in tenant
type MemberRole string

const (
	RoleOwner  MemberRole = "OWNER"
	RoleAdmin  MemberRole = "ADMIN"
	RoleMember MemberRole = "MEMBER"
	RoleViewer MemberRole = "VIEWER"
)

// MemberStatus in tenant
type MemberStatus string

const (
	MemberActive      MemberStatus = "ACTIVE"
	MemberInvited     MemberStatus = "INVITED"
	MemberSuspended   MemberStatus = "SUSPENDED"
	MemberDeactivated MemberStatus = "DEACTIVATED"
)

// TenantResource represents resource scoped to tenant
type TenantResource struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenantId"` // Required for all resources
	Kind      string                 `json:"kind"`     // Resource type
	Name      string                 `json:"name"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}

// TenantContext passed through requests
type TenantContext struct {
	TenantID       string
	UserID         string
	Role           MemberRole
	Quota          *TenantQuota
	IsolationLevel TenantIsolation
}

// TenantConfig configures multi-tenancy behavior
type TenantConfig struct {
	Enabled           bool
	IsolationLevel    TenantIsolation
	AllowSubdomains   bool // api-{tenant}.example.com
	AllowCustomDomain bool // custom.example.com
	QuotaEnforcement  bool
	SoftDeleteTenants bool
	DataEncryption    bool // Per-tenant encryption key
}
