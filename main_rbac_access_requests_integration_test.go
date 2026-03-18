package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"example.com/axiomnizam/internal/auth"
	"example.com/axiomnizam/internal/rbac"
	"github.com/gin-gonic/gin"
)

func TestMainRouteProtectionPhase2(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}

	src := string(content)
	required := []string{
		`router.Any("/api/custom", authMiddleware, apiBuilderHandler.InvokeCustomAPI)`,
		`router.Any("/api/custom/*path", authMiddleware, apiBuilderHandler.InvokeCustomAPI)`,
		`builderAPI.DELETE("/csv/uploads/:id", adminOrSysMiddleware, apiBuilderHandler.DeleteCSVUpload)`,
		`builderAPI.DELETE("/dashboards/:id", adminOrSysMiddleware, apiBuilderHandler.DeleteDashboard)`,
		`builderAPI.POST("/scanner/scan", adminOrSysMiddleware, apiBuilderHandler.ScanFile)`,
	}

	for _, snippet := range required {
		if !strings.Contains(src, snippet) {
			t.Fatalf("missing expected protected route mapping: %s", snippet)
		}
	}
}

const (
	testAdminToken = "admin-token"
	testUserToken  = "user-token"
)

type mainRBACManagerAdapter struct {
	core *rbac.InMemoryRBACManager
}

func (m *mainRBACManagerAdapter) CreateRole(role *rbac.Role) (*rbac.Role, error) {
	return m.core.CreateRole(role)
}

func (m *mainRBACManagerAdapter) GetRole(id string) (*rbac.Role, error) {
	return m.core.GetRole(id)
}

func (m *mainRBACManagerAdapter) ListRoles(tenantID string) ([]*rbac.Role, error) {
	return m.core.ListRoles(tenantID)
}

func (m *mainRBACManagerAdapter) UpdateRole(role *rbac.Role) (*rbac.Role, error) {
	return m.core.UpdateRole(role)
}

func (m *mainRBACManagerAdapter) DeleteRole(id string) error {
	return m.core.DeleteRole(id)
}

func (m *mainRBACManagerAdapter) CreateRoleBinding(binding *rbac.RoleBinding) (*rbac.RoleBinding, error) {
	return m.core.CreateRoleBinding(binding)
}

func (m *mainRBACManagerAdapter) ListRoleBindings(tenantID, principalID string) ([]*rbac.RoleBinding, error) {
	items, err := m.core.ListRoleBindings("", principalID)
	if err != nil {
		return nil, err
	}
	if tenantID == "" {
		return items, nil
	}

	filtered := make([]*rbac.RoleBinding, 0, len(items))
	for _, item := range items {
		if item.TenantID == tenantID {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func (m *mainRBACManagerAdapter) DeleteRoleBinding(id string) error {
	return m.core.DeleteRoleBinding(id)
}

func (m *mainRBACManagerAdapter) CheckPermission(req *rbac.PermissionCheck) (*rbac.PermissionCheckResult, error) {
	allowed, err := m.core.CheckPermission(req.PrincipalID, req.Resource, req.Action)
	if err != nil {
		return nil, err
	}
	reason := "denied"
	if allowed {
		reason = "allowed"
	}
	return &rbac.PermissionCheckResult{Allowed: allowed, Reason: reason}, nil
}

func (m *mainRBACManagerAdapter) ListPermissions(tenantID, resource string) ([]*rbac.Permission, error) {
	items, err := m.core.ListPermissions("")
	if err != nil {
		return nil, err
	}
	filtered := make([]*rbac.Permission, 0, len(items))
	for _, item := range items {
		if tenantID != "" && item.TenantID != tenantID {
			continue
		}
		if resource != "" && item.Resource != resource {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, nil
}

func (m *mainRBACManagerAdapter) CreateAccessRequest(req *rbac.AccessRequest) (*rbac.AccessRequest, error) {
	return m.core.CreateAccessRequest(req)
}

func (m *mainRBACManagerAdapter) ListAccessRequests(tenantID, principalID, status string) ([]*rbac.AccessRequest, error) {
	return m.core.ListAccessRequests(tenantID, principalID, status)
}

func (m *mainRBACManagerAdapter) ApproveAccessRequest(id, approvedBy string) (*rbac.AccessRequest, error) {
	return m.core.ApproveAccessRequest(id, approvedBy)
}

func (m *mainRBACManagerAdapter) RejectAccessRequest(id, rejectedBy, reason string) (*rbac.AccessRequest, error) {
	return m.core.RejectAccessRequest(id, rejectedBy, reason)
}

func claimsForTestToken(token string) *auth.Claims {
	switch token {
	case testAdminToken:
		return &auth.Claims{PreferredUsername: "admin-user", RealmAccess: auth.RealmAccess{Roles: []string{"admin"}}}
	case testUserToken:
		return &auth.Claims{PreferredUsername: "normal-user", RealmAccess: auth.RealmAccess{Roles: []string{"user"}}}
	default:
		return nil
	}
}

func requireTestAuthentication(c *gin.Context) (*auth.Claims, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		c.Abort()
		return nil, false
	}

	token, err := auth.ExtractBearerToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid authorization header: %v", err)})
		c.Abort()
		return nil, false
	}

	claims := claimsForTestToken(token)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return nil, false
	}

	c.Set("user", claims)
	c.Set("username", claims.PreferredUsername)
	c.Set("roles", claims.RealmAccess.Roles)
	c.Set("token", token)
	return claims, true
}

func hasMainRBACPrivilegedRole(claims *auth.Claims) bool {
	return claims != nil &&
		(claims.HasRole("admin") || claims.HasRole("system-manager") || claims.HasRole("sysadmin") || claims.HasRole("system_admin") || claims.HasRole("system-admin"))
}

func setupMainLikeRBACRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	manager := &mainRBACManagerAdapter{core: rbac.NewInMemoryRBACManager()}
	rbacHandler := rbac.NewRBACHandler(manager)

	authMiddleware := func(c *gin.Context) {
		if _, ok := requireTestAuthentication(c); !ok {
			return
		}
		c.Next()
	}

	adminOrSysMiddleware := func(c *gin.Context) {
		claims, ok := requireTestAuthentication(c)
		if !ok {
			return
		}

		if !hasMainRBACPrivilegedRole(claims) {
			roles := []string{}
			roles = claims.RealmAccess.Roles
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "forbidden: user must have one of roles [admin system-manager sysadmin system_admin system-admin]",
				"user_roles": roles,
				"required":   []string{"admin", "system-manager", "sysadmin", "system_admin", "system-admin"},
			})
			c.Abort()
			return
		}
		c.Next()
	}

	rbacAPI := router.Group("/api/v1/rbac", authMiddleware)
	{
		rbacAPI.POST("/access-requests", rbacHandler.CreateAccessRequest)
		rbacAPI.GET("/access-requests", rbacHandler.ListAccessRequests)
		rbacAPI.POST("/access-requests/:id/approve", adminOrSysMiddleware, rbacHandler.ApproveAccessRequest)
		rbacAPI.POST("/access-requests/:id/reject", adminOrSysMiddleware, rbacHandler.RejectAccessRequest)
	}

	return router
}

func performMainRBACRequest(t *testing.T, router *gin.Engine, method, path, bearerToken string, payload interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var body *bytes.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		body = bytes.NewReader(encoded)
	} else {
		body = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func decodeMainRBACResponse(t *testing.T, rr *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(rr.Body.Bytes(), target); err != nil {
		t.Fatalf("failed to decode JSON response: %v; body=%s", err, rr.Body.String())
	}
}

func assertMainRBACUnauthorizedCreate(t *testing.T, router *gin.Engine) {
	t.Helper()
	resp := performMainRBACRequest(t, router, http.MethodPost, "/api/v1/rbac/access-requests", "", map[string]interface{}{
		"tenantId":     "tenant-main",
		"principalId":  "user-main",
		"resourceType": "ROLE",
		"action":       "READ",
	})
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", resp.Code, resp.Body.String())
	}
}

func assertMainRBACAuthenticatedLifecycleFlow(t *testing.T, router *gin.Engine) {
	t.Helper()

	createResp := performMainRBACRequest(t, router, http.MethodPost, "/api/v1/rbac/access-requests", testUserToken, map[string]interface{}{
		"tenantId":      "tenant-main",
		"principalId":   "user-main",
		"resourceType":  "ROLE",
		"resourceId":    "role-reader",
		"action":        "READ",
		"justification": "need temporary reader access",
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create 201, got %d: %s", createResp.Code, createResp.Body.String())
	}

	var created rbac.AccessRequest
	decodeMainRBACResponse(t, createResp, &created)
	if created.ID == "" {
		t.Fatal("expected access request ID")
	}

	listResp := performMainRBACRequest(t, router, http.MethodGet, "/api/v1/rbac/access-requests?tenantId=tenant-main&principalId=user-main&status=PENDING", testUserToken, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d: %s", listResp.Code, listResp.Body.String())
	}
	var listed struct {
		AccessRequests []rbac.AccessRequest `json:"accessRequests"`
		Count          int                  `json:"count"`
	}
	decodeMainRBACResponse(t, listResp, &listed)
	if listed.Count != 1 || len(listed.AccessRequests) != 1 {
		t.Fatalf("expected exactly one pending request, got count=%d body=%s", listed.Count, listResp.Body.String())
	}

	approveForbidden := performMainRBACRequest(t, router, http.MethodPost, "/api/v1/rbac/access-requests/"+created.ID+"/approve", testUserToken, map[string]interface{}{
		"approvedBy": "normal-user",
	})
	if approveForbidden.Code != http.StatusForbidden {
		t.Fatalf("expected approve 403 for non-admin user, got %d: %s", approveForbidden.Code, approveForbidden.Body.String())
	}

	approveOK := performMainRBACRequest(t, router, http.MethodPost, "/api/v1/rbac/access-requests/"+created.ID+"/approve", testAdminToken, map[string]interface{}{
		"approvedBy": "admin-user",
	})
	if approveOK.Code != http.StatusOK {
		t.Fatalf("expected approve 200 for admin user, got %d: %s", approveOK.Code, approveOK.Body.String())
	}

	var approved rbac.AccessRequest
	decodeMainRBACResponse(t, approveOK, &approved)
	if approved.Status != rbac.RequestStatusApproved {
		t.Fatalf("expected APPROVED status, got %s", approved.Status)
	}
}

func TestMainRouterRBACAccessRequestsAuthIntegration(t *testing.T) {
	router := setupMainLikeRBACRouter()

	t.Run("unauthorized request is rejected by auth middleware", func(t *testing.T) {
		assertMainRBACUnauthorizedCreate(t, router)
	})

	t.Run("authenticated create/list with admin gate on approve", func(t *testing.T) {
		assertMainRBACAuthenticatedLifecycleFlow(t, router)
	})
}
