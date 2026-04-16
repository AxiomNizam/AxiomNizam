package access

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"example.com/axiomnizam/internal/iam/token"
	"example.com/axiomnizam/internal/storage/events"
	"example.com/axiomnizam/internal/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Controller manages IAM-integrated access control, access keys, and bucket sharing.
type Controller struct {
	mu         sync.RWMutex
	policies   map[string]*models.TenantPolicy // key = tenantID/userID/bucket
	accessKeys map[string]*models.AccessKey    // key = accessKeyID
	shares     map[string]*models.BucketShare  // key = shareID
	auditLog   *events.AuditLog
}

// NewController creates an access controller.
func NewController(auditLog *events.AuditLog) *Controller {
	return &Controller{
		policies:   make(map[string]*models.TenantPolicy),
		accessKeys: make(map[string]*models.AccessKey),
		shares:     make(map[string]*models.BucketShare),
		auditLog:   auditLog,
	}
}

// ---------------------------------------------------------------------------
// IAM Middleware — Extract user context from JWT
// ---------------------------------------------------------------------------

// StorageContext holds the authenticated user information extracted from JWT.
type StorageContext struct {
	UserID   string
	Email    string
	TenantID string
	Roles    []string
}

// GetStorageContext extracts the storage context from a Gin context.
// Requires IAM JWTAuth middleware to have run first.
func GetStorageContext(c *gin.Context) *StorageContext {
	v, exists := c.Get("storage_context")
	if !exists {
		return nil
	}
	sc, _ := v.(*StorageContext)
	return sc
}

// RequireStorageAuth is middleware that extracts the IAM user context and
// maps it to a storage-level identity. It also supports access key authentication.
func (ac *Controller) RequireStorageAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Option 1: Check for storage access key in X-Storage-Access-Key header.
		accessKeyID := c.GetHeader("X-Storage-Access-Key")
		secretKey := c.GetHeader("X-Storage-Secret-Key")
		if accessKeyID != "" && secretKey != "" {
			ak := ac.ValidateAccessKey(accessKeyID, secretKey)
			if ak == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid storage access key"})
				return
			}
			sc := &StorageContext{
				UserID:   ak.UserID,
				TenantID: ak.TenantID,
				Roles:    []string{string(ak.Role)},
			}
			c.Set("storage_context", sc)
			c.Set("storage_access_key", ak)
			c.Next()
			return
		}

		// Option 2: Use IAM JWT claims from IAM middleware.
		claimsRaw, exists := c.Get("iam_claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication required",
				"details": "provide IAM JWT token or storage access key",
			})
			return
		}

		claims, ok := claimsRaw.(*token.IAMClaims)
		if !ok || claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authentication claims"})
			return
		}

		// Derive tenant ID: use the user's sub (user ID) as default tenant.
		// Applications may override with X-Storage-Tenant header.
		tenantID := claims.Sub
		if override := strings.TrimSpace(c.GetHeader("X-Storage-Tenant")); override != "" {
			// Only sysadmin/admin can impersonate tenants.
			if hasRole(claims.Roles, "sysadmin", "system-manager", "admin") {
				tenantID = override
			}
		}

		sc := &StorageContext{
			UserID:   claims.Sub,
			Email:    claims.Email,
			TenantID: tenantID,
			Roles:    claims.Roles,
		}
		c.Set("storage_context", sc)
		c.Next()
	}
}

// RequireStorageRole checks that the authenticated user holds one of the specified
// storage roles (either via IAM role mapping or direct storage policy).
func (ac *Controller) RequireStorageRole(allowed ...models.StorageRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		sc := GetStorageContext(c)
		if sc == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		// Sysadmin / system-manager bypass all storage role checks.
		if hasRole(sc.Roles, "sysadmin", "system-manager") {
			c.Next()
			return
		}

		// Map IAM roles to storage roles and check.
		for _, iamRole := range sc.Roles {
			storageRole := MapIAMRoleToStorageRole(iamRole)
			for _, a := range allowed {
				if storageRole == a {
					c.Next()
					return
				}
			}
		}

		// Check direct storage access key role.
		for _, r := range sc.Roles {
			for _, a := range allowed {
				if models.StorageRole(r) == a {
					c.Next()
					return
				}
			}
		}

		// Check bucket-specific policies.
		bucket := c.Param("bucket")
		if bucket != "" {
			for _, a := range allowed {
				if ac.HasBucketAccess(sc.UserID, sc.TenantID, bucket, a) {
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":    "insufficient storage permissions",
			"required": allowed,
		})
	}
}

// RequireBucketAccess checks that the authenticated user can access the specific
// bucket named in the :bucket path parameter. Checks ownership, shares, and policies.
func (ac *Controller) RequireBucketAccess(minRole models.StorageRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		sc := GetStorageContext(c)
		if sc == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		// Sysadmin bypass.
		if hasRole(sc.Roles, "sysadmin", "system-manager") {
			c.Next()
			return
		}

		bucket := c.Param("bucket")
		if bucket == "" {
			c.Next()
			return
		}

		// Check if access key has bucket scope.
		if akRaw, exists := c.Get("storage_access_key"); exists {
			ak := akRaw.(*models.AccessKey)
			if len(ak.BucketScope) > 0 {
				found := false
				for _, b := range ak.BucketScope {
					if b == bucket {
						found = true
						break
					}
				}
				if !found {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"error":  "access key does not have access to this bucket",
						"bucket": bucket,
					})
					return
				}
			}
		}

		// Check direct IAM role mapping.
		for _, iamRole := range sc.Roles {
			sr := MapIAMRoleToStorageRole(iamRole)
			if roleAtLeast(sr, minRole) {
				c.Next()
				return
			}
		}

		// Check bucket-level policy or share.
		if ac.HasBucketAccess(sc.UserID, sc.TenantID, bucket, minRole) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":    "insufficient permissions for bucket",
			"bucket":   bucket,
			"required": minRole,
		})
	}
}

// ---------------------------------------------------------------------------
// IAM-to-Storage Role Mapping
// ---------------------------------------------------------------------------

// MapIAMRoleToStorageRole maps an IAM role to the equivalent storage role.
func MapIAMRoleToStorageRole(iamRole string) models.StorageRole {
	switch strings.ToLower(strings.TrimSpace(iamRole)) {
	case "sysadmin", "system-manager", "system_admin", "system-admin":
		return models.StorageRoleAdmin
	case "admin":
		return models.StorageRoleAdmin
	case "manager":
		return models.StorageRoleWriter
	case "user":
		return models.StorageRoleReader
	case "uploader":
		return models.StorageRoleUploader
	default:
		return models.StorageRoleReader
	}
}

// roleAtLeast checks if `have` is at least as permissive as `need`.
func roleAtLeast(have, need models.StorageRole) bool {
	order := map[models.StorageRole]int{
		models.StorageRoleAdmin:    4,
		models.StorageRoleWriter:   3,
		models.StorageRoleBrowser:  2,
		models.StorageRoleReader:   1,
		models.StorageRoleUploader: 1,
	}
	return order[have] >= order[need]
}

func hasRole(roles []string, targets ...string) bool {
	for _, r := range roles {
		nr := strings.ToLower(strings.TrimSpace(r))
		for _, t := range targets {
			if nr == t {
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Access Keys (Service Accounts for Applications)
// ---------------------------------------------------------------------------

// CreateAccessKey generates a new storage access key bound to a user.
func (ac *Controller) CreateAccessKey(userID, tenantID, name, description string, role models.StorageRole, bucketScope []string, expiresAt *time.Time) (*models.AccessKey, error) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	accessKeyID := "AXAK" + generateRandomHex(16)
	secretAccessKey := generateRandomHex(32)

	ak := &models.AccessKey{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Name:            name,
		Description:     description,
		UserID:          userID,
		TenantID:        tenantID,
		Role:            role,
		BucketScope:     bucketScope,
		Active:          true,
		ExpiresAt:       expiresAt,
		CreatedAt:       time.Now().UTC(),
	}

	ac.accessKeys[accessKeyID] = ak

	if ac.auditLog != nil {
		ac.auditLog.Record(models.StorageEvent{
			Type:     "accesskey.created",
			TenantID: tenantID,
			UserID:   userID,
			Details:  fmt.Sprintf("access key %s created for %s (role: %s)", accessKeyID, name, role),
		})
	}

	return ak, nil
}

// ValidateAccessKey verifies an access key ID + secret pair.
func (ac *Controller) ValidateAccessKey(accessKeyID, secretKey string) *models.AccessKey {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	ak, exists := ac.accessKeys[accessKeyID]
	if !exists || !ak.Active {
		return nil
	}

	// Check expiration.
	if ak.ExpiresAt != nil && time.Now().After(*ak.ExpiresAt) {
		return nil
	}

	// Verify secret using HMAC comparison.
	if !hmac.Equal([]byte(ak.SecretAccessKey), []byte(secretKey)) {
		return nil
	}

	// Update last used.
	now := time.Now().UTC()
	ak.LastUsedAt = &now

	return ak
}

// ListAccessKeys returns all access keys for a user.
func (ac *Controller) ListAccessKeys(userID, tenantID string) []*models.AccessKey {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	var result []*models.AccessKey
	for _, ak := range ac.accessKeys {
		if (userID == "" || ak.UserID == userID) && (tenantID == "" || ak.TenantID == tenantID) {
			// Never return the secret key in list operations.
			safe := *ak
			safe.SecretAccessKey = ""
			result = append(result, &safe)
		}
	}
	return result
}

// RevokeAccessKey deactivates an access key.
func (ac *Controller) RevokeAccessKey(accessKeyID, userID string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ak, exists := ac.accessKeys[accessKeyID]
	if !exists {
		return fmt.Errorf("access key %q not found", accessKeyID)
	}
	// Users can only revoke their own keys; admins bypass via the handler.
	if ak.UserID != userID && userID != "" {
		return fmt.Errorf("access key %q does not belong to user %q", accessKeyID, userID)
	}

	ak.Active = false

	if ac.auditLog != nil {
		ac.auditLog.Record(models.StorageEvent{
			Type:     "accesskey.revoked",
			TenantID: ak.TenantID,
			UserID:   userID,
			Details:  fmt.Sprintf("access key %s revoked", accessKeyID),
		})
	}

	return nil
}

// DeleteAccessKey permanently removes an access key.
func (ac *Controller) DeleteAccessKey(accessKeyID, userID string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ak, exists := ac.accessKeys[accessKeyID]
	if !exists {
		return fmt.Errorf("access key %q not found", accessKeyID)
	}
	if ak.UserID != userID && userID != "" {
		return fmt.Errorf("access key %q does not belong to user %q", accessKeyID, userID)
	}

	delete(ac.accessKeys, accessKeyID)
	return nil
}

// ActiveAccessKeyCount returns the number of active access keys.
func (ac *Controller) ActiveAccessKeyCount() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	count := 0
	for _, ak := range ac.accessKeys {
		if ak.Active {
			count++
		}
	}
	return count
}

// ---------------------------------------------------------------------------
// Bucket Sharing
// ---------------------------------------------------------------------------

// ShareBucket creates a share granting access to a bucket.
func (ac *Controller) ShareBucket(share models.BucketShare) (*models.BucketShare, error) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	share.ID = uuid.New().String()
	share.SharedAt = time.Now().UTC()
	share.Active = true

	ac.shares[share.ID] = &share

	if ac.auditLog != nil {
		ac.auditLog.Record(models.StorageEvent{
			Type:     "bucket.shared",
			TenantID: share.TenantID,
			Bucket:   share.BucketName,
			UserID:   share.SharedBy,
			Details:  fmt.Sprintf("shared with %s (%s) as %s", share.GranteeID, share.GranteeType, share.Role),
		})
	}

	return &share, nil
}

// ListBucketShares returns all shares for a bucket.
func (ac *Controller) ListBucketShares(tenantID, bucket string) []*models.BucketShare {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	var result []*models.BucketShare
	for _, s := range ac.shares {
		if (tenantID == "" || s.TenantID == tenantID) && (bucket == "" || s.BucketName == bucket) {
			result = append(result, s)
		}
	}
	return result
}

// ListUserShares returns all shares granted to a specific user.
func (ac *Controller) ListUserShares(userID string) []*models.BucketShare {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	var result []*models.BucketShare
	for _, s := range ac.shares {
		if s.GranteeID == userID && s.Active {
			result = append(result, s)
		}
	}
	return result
}

// RevokeShare deactivates a bucket share.
func (ac *Controller) RevokeShare(shareID, userID string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	s, exists := ac.shares[shareID]
	if !exists {
		return fmt.Errorf("share %q not found", shareID)
	}
	s.Active = false

	if ac.auditLog != nil {
		ac.auditLog.Record(models.StorageEvent{
			Type:     "bucket.share.revoked",
			TenantID: s.TenantID,
			Bucket:   s.BucketName,
			UserID:   userID,
			Details:  fmt.Sprintf("share %s revoked for %s", shareID, s.GranteeID),
		})
	}

	return nil
}

// DeleteShare permanently removes a share.
func (ac *Controller) DeleteShare(shareID string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if _, exists := ac.shares[shareID]; !exists {
		return fmt.Errorf("share %q not found", shareID)
	}
	delete(ac.shares, shareID)
	return nil
}

// ActiveShareCount returns the number of active shares.
func (ac *Controller) ActiveShareCount() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	count := 0
	for _, s := range ac.shares {
		if s.Active {
			count++
		}
	}
	return count
}

// ---------------------------------------------------------------------------
// Bucket Access Policies
// ---------------------------------------------------------------------------

// SetPolicy creates or updates a bucket access policy.
func (ac *Controller) SetPolicy(p models.TenantPolicy) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	key := policyKey(p.TenantID, p.UserID, p.BucketName)
	p.GrantedAt = time.Now().UTC()
	ac.policies[key] = &p
	return nil
}

// GetPolicy returns a specific policy.
func (ac *Controller) GetPolicy(tenantID, userID, bucket string) *models.TenantPolicy {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.policies[policyKey(tenantID, userID, bucket)]
}

// ListPolicies returns all policies, optionally filtered.
func (ac *Controller) ListPolicies(tenantID string) []*models.TenantPolicy {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	var result []*models.TenantPolicy
	for _, p := range ac.policies {
		if tenantID == "" || p.TenantID == tenantID {
			result = append(result, p)
		}
	}
	return result
}

// DeletePolicy removes a policy.
func (ac *Controller) DeletePolicy(tenantID, userID, bucket string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	key := policyKey(tenantID, userID, bucket)
	if _, exists := ac.policies[key]; !exists {
		return fmt.Errorf("policy not found")
	}
	delete(ac.policies, key)
	return nil
}

// PolicyCount returns total policies.
func (ac *Controller) PolicyCount() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return len(ac.policies)
}

// HasBucketAccess checks if a user has at least the given role on a bucket
// through policies or shares.
func (ac *Controller) HasBucketAccess(userID, tenantID, bucket string, minRole models.StorageRole) bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Check direct policy.
	if p := ac.policies[policyKey(tenantID, userID, bucket)]; p != nil {
		if roleAtLeast(p.Role, minRole) {
			if p.ExpiresAt == nil || time.Now().Before(*p.ExpiresAt) {
				return true
			}
		}
	}

	// Check wildcard policy (all buckets).
	if p := ac.policies[policyKey(tenantID, userID, "*")]; p != nil {
		if roleAtLeast(p.Role, minRole) {
			if p.ExpiresAt == nil || time.Now().Before(*p.ExpiresAt) {
				return true
			}
		}
	}

	// Check bucket shares.
	for _, s := range ac.shares {
		if s.GranteeID == userID && s.BucketName == bucket && s.Active {
			if roleAtLeast(s.Role, minRole) {
				if s.ExpiresAt == nil || time.Now().Before(*s.ExpiresAt) {
					return true
				}
			}
		}
	}

	return false
}

func policyKey(tenantID, userID, bucket string) string {
	return tenantID + "/" + userID + "/" + bucket
}

func generateRandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
