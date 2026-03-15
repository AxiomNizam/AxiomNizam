package rbac

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type accessRequestListResponse struct {
	AccessRequests []AccessRequest `json:"accessRequests"`
	Count          int             `json:"count"`
}

type handlerTestRBACManager struct {
	core *InMemoryRBACManager
}

func (m *handlerTestRBACManager) CreateRole(role *Role) (*Role, error) {
	return m.core.CreateRole(role)
}

func (m *handlerTestRBACManager) GetRole(id string) (*Role, error) {
	return m.core.GetRole(id)
}

func (m *handlerTestRBACManager) ListRoles(tenantID string) ([]*Role, error) {
	return m.core.ListRoles(tenantID)
}

func (m *handlerTestRBACManager) UpdateRole(role *Role) (*Role, error) {
	return m.core.UpdateRole(role)
}

func (m *handlerTestRBACManager) DeleteRole(id string) error {
	return m.core.DeleteRole(id)
}

func (m *handlerTestRBACManager) CreateRoleBinding(binding *RoleBinding) (*RoleBinding, error) {
	return m.core.CreateRoleBinding(binding)
}

func (m *handlerTestRBACManager) ListRoleBindings(tenantID, principalID string) ([]*RoleBinding, error) {
	items, err := m.core.ListRoleBindings("", principalID)
	if err != nil {
		return nil, err
	}
	if tenantID == "" {
		return items, nil
	}
	filtered := make([]*RoleBinding, 0, len(items))
	for _, item := range items {
		if item.TenantID == tenantID {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func (m *handlerTestRBACManager) DeleteRoleBinding(id string) error {
	return m.core.DeleteRoleBinding(id)
}

func (m *handlerTestRBACManager) CheckPermission(req *PermissionCheck) (*PermissionCheckResult, error) {
	allowed, err := m.core.CheckPermission(req.PrincipalID, req.Resource, req.Action)
	if err != nil {
		return nil, err
	}
	result := &PermissionCheckResult{Allowed: allowed}
	if allowed {
		result.Reason = "allowed"
	} else {
		result.Reason = "denied"
	}
	return result, nil
}

func (m *handlerTestRBACManager) ListPermissions(tenantID, resource string) ([]*Permission, error) {
	items, err := m.core.ListPermissions("")
	if err != nil {
		return nil, err
	}
	filtered := make([]*Permission, 0, len(items))
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

func (m *handlerTestRBACManager) CreateAccessRequest(req *AccessRequest) (*AccessRequest, error) {
	return m.core.CreateAccessRequest(req)
}

func (m *handlerTestRBACManager) ListAccessRequests(tenantID, principalID, status string) ([]*AccessRequest, error) {
	return m.core.ListAccessRequests(tenantID, principalID, status)
}

func (m *handlerTestRBACManager) ApproveAccessRequest(id, approvedBy string) (*AccessRequest, error) {
	return m.core.ApproveAccessRequest(id, approvedBy)
}

func (m *handlerTestRBACManager) RejectAccessRequest(id, rejectedBy, reason string) (*AccessRequest, error) {
	return m.core.RejectAccessRequest(id, rejectedBy, reason)
}

func setupRBACAccessRequestTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	manager := &handlerTestRBACManager{core: NewInMemoryRBACManager()}
	handler := NewRBACHandler(manager)
	router := gin.New()

	rbacAPI := router.Group("/api/v1/rbac")
	rbacAPI.POST("/access-requests", handler.CreateAccessRequest)
	rbacAPI.GET("/access-requests", handler.ListAccessRequests)
	rbacAPI.POST("/access-requests/:id/approve", handler.ApproveAccessRequest)
	rbacAPI.POST("/access-requests/:id/reject", handler.RejectAccessRequest)

	return router
}

func performJSONRequest(t *testing.T, router *gin.Engine, method, path string, payload interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var body *bytes.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal request payload: %v", err)
		}
		body = bytes.NewReader(encoded)
	} else {
		body = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func decodeJSONBody(t *testing.T, rr *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(rr.Body.Bytes(), target); err != nil {
		t.Fatalf("failed to decode response body: %v; body=%s", err, rr.Body.String())
	}
}

func TestAccessRequestLifecycleAPI(t *testing.T) {
	t.Run("create and list pending access request", func(t *testing.T) {
		router := setupRBACAccessRequestTestRouter()

		createResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/rbac/access-requests", map[string]interface{}{
			"tenantId":      "tenant-a",
			"principalId":   "user-a",
			"resourceType":  "ROLE",
			"resourceId":    "role-reader",
			"action":        "READ",
			"duration":      120,
			"justification": "Need read access",
		})

		if createResp.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", createResp.Code, createResp.Body.String())
		}

		var created AccessRequest
		decodeJSONBody(t, createResp, &created)
		if created.ID == "" {
			t.Fatal("expected created access request ID")
		}
		if created.Status != RequestStatusPending {
			t.Fatalf("expected PENDING status, got %s", created.Status)
		}
		if created.ExpiresAt.IsZero() {
			t.Fatal("expected non-zero expiresAt when duration is provided")
		}

		listResp := performJSONRequest(t, router, http.MethodGet, "/api/v1/rbac/access-requests?tenantId=tenant-a&principalId=user-a&status=PENDING", nil)
		if listResp.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", listResp.Code, listResp.Body.String())
		}

		var listed accessRequestListResponse
		decodeJSONBody(t, listResp, &listed)
		if listed.Count != 1 {
			t.Fatalf("expected count=1, got %d", listed.Count)
		}
		if len(listed.AccessRequests) != 1 || listed.AccessRequests[0].ID != created.ID {
			t.Fatalf("expected listed request ID %s, got %#v", created.ID, listed.AccessRequests)
		}
	})

	t.Run("approve transition", func(t *testing.T) {
		router := setupRBACAccessRequestTestRouter()

		createResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/rbac/access-requests", map[string]interface{}{
			"tenantId":     "tenant-b",
			"principalId":  "user-b",
			"resourceType": "ROLE",
			"resourceId":   "role-admin",
			"action":       "WRITE",
		})
		if createResp.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", createResp.Code, createResp.Body.String())
		}

		var created AccessRequest
		decodeJSONBody(t, createResp, &created)

		approveResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/rbac/access-requests/"+created.ID+"/approve", map[string]interface{}{
			"approvedBy": "admin-1",
		})
		if approveResp.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", approveResp.Code, approveResp.Body.String())
		}

		var approved AccessRequest
		decodeJSONBody(t, approveResp, &approved)
		if approved.Status != RequestStatusApproved {
			t.Fatalf("expected APPROVED status, got %s", approved.Status)
		}
		if approved.ApprovedBy != "admin-1" {
			t.Fatalf("expected approvedBy=admin-1, got %s", approved.ApprovedBy)
		}
		if approved.ApprovedAt.IsZero() {
			t.Fatal("expected approvedAt to be set")
		}
	})

	t.Run("reject transition", func(t *testing.T) {
		router := setupRBACAccessRequestTestRouter()

		createResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/rbac/access-requests", map[string]interface{}{
			"tenantId":     "tenant-c",
			"principalId":  "user-c",
			"resourceType": "ROLE",
			"resourceId":   "role-editor",
			"action":       "READ",
		})
		if createResp.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", createResp.Code, createResp.Body.String())
		}

		var created AccessRequest
		decodeJSONBody(t, createResp, &created)

		rejectResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/rbac/access-requests/"+created.ID+"/reject", map[string]interface{}{
			"rejectedBy": "admin-2",
			"reason":     "Policy violation",
		})
		if rejectResp.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rejectResp.Code, rejectResp.Body.String())
		}

		var rejected AccessRequest
		decodeJSONBody(t, rejectResp, &rejected)
		if rejected.Status != RequestStatusRejected {
			t.Fatalf("expected REJECTED status, got %s", rejected.Status)
		}
		if rejected.RejectionReason != "Policy violation" {
			t.Fatalf("expected rejection reason to be set, got %q", rejected.RejectionReason)
		}
		if rejected.RejectedAt.IsZero() {
			t.Fatal("expected rejectedAt to be set")
		}
		if got, ok := rejected.Metadata["rejectedBy"].(string); !ok || got != "admin-2" {
			t.Fatalf("expected metadata.rejectedBy=admin-2, got %#v", rejected.Metadata["rejectedBy"])
		}
	})

	t.Run("expiry transition blocks approval and appears as expired in list", func(t *testing.T) {
		router := setupRBACAccessRequestTestRouter()

		createResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/rbac/access-requests", map[string]interface{}{
			"tenantId":     "tenant-d",
			"principalId":  "user-d",
			"resourceType": "ROLE",
			"resourceId":   "role-temp",
			"action":       "READ",
			"duration":     1,
		})
		if createResp.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", createResp.Code, createResp.Body.String())
		}

		var created AccessRequest
		decodeJSONBody(t, createResp, &created)

		time.Sleep(1200 * time.Millisecond)

		approveResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/rbac/access-requests/"+created.ID+"/approve", map[string]interface{}{
			"approvedBy": "admin-3",
		})
		if approveResp.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500 for expired request, got %d: %s", approveResp.Code, approveResp.Body.String())
		}

		var approveErr map[string]interface{}
		decodeJSONBody(t, approveResp, &approveErr)
		errText, _ := approveErr["error"].(string)
		if !strings.Contains(strings.ToLower(errText), "expired") {
			t.Fatalf("expected expired error, got %q", errText)
		}

		listResp := performJSONRequest(t, router, http.MethodGet, "/api/v1/rbac/access-requests?tenantId=tenant-d&principalId=user-d&status=EXPIRED", nil)
		if listResp.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", listResp.Code, listResp.Body.String())
		}

		var listed accessRequestListResponse
		decodeJSONBody(t, listResp, &listed)
		if listed.Count != 1 {
			t.Fatalf("expected one expired request, got %d", listed.Count)
		}
		if len(listed.AccessRequests) != 1 || listed.AccessRequests[0].ID != created.ID {
			t.Fatalf("expected expired request ID %s, got %#v", created.ID, listed.AccessRequests)
		}
		if listed.AccessRequests[0].Status != RequestStatusExpired {
			t.Fatalf("expected EXPIRED status, got %s", listed.AccessRequests[0].Status)
		}
	})
}
