package admin

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/axiomnizam/internal/iam/authn"
	"example.com/axiomnizam/internal/iam/authz"
	"example.com/axiomnizam/internal/iam/identity"
	iammw "example.com/axiomnizam/internal/iam/middleware"
	"example.com/axiomnizam/internal/iam/oauth"
	"example.com/axiomnizam/internal/iam/storage"
	"example.com/axiomnizam/internal/iam/token"
	"github.com/gin-gonic/gin"
)

// Handler bundles all sysadmin (master-realm) API endpoints.
type Handler struct {
	users        *storage.PostgresUserRepository
	clients      *storage.EtcdClientRepository
	roles        *storage.EtcdRoleRepository
	bindings     *storage.EtcdRoleBindingRepository
	sessions     *storage.EtcdSessionRepository
	refreshRepo  *storage.EtcdRefreshTokenRepository
	revokedStore *storage.EtcdRevokedTokenStore
	codeRepo     *storage.EtcdCodeRepository
	authorizer   *authz.Authorizer
	issuer       *token.Issuer
	authn        *authn.Authenticator
}

// NewHandler creates the admin handler.
func NewHandler(
	users *storage.PostgresUserRepository,
	clients *storage.EtcdClientRepository,
	roles *storage.EtcdRoleRepository,
	bindings *storage.EtcdRoleBindingRepository,
	sessions *storage.EtcdSessionRepository,
	refreshRepo *storage.EtcdRefreshTokenRepository,
	revokedStore *storage.EtcdRevokedTokenStore,
	codeRepo *storage.EtcdCodeRepository,
	authorizer *authz.Authorizer,
	issuer *token.Issuer,
	authenticator *authn.Authenticator,
) *Handler {
	return &Handler{
		users:        users,
		clients:      clients,
		roles:        roles,
		bindings:     bindings,
		sessions:     sessions,
		refreshRepo:  refreshRepo,
		revokedStore: revokedStore,
		codeRepo:     codeRepo,
		authorizer:   authorizer,
		issuer:       issuer,
		authn:        authenticator,
	}
}

func normalizeRoleNames(roleNames []string) []string {
	seen := make(map[string]struct{}, len(roleNames))
	out := make([]string, 0, len(roleNames))

	for _, roleName := range roleNames {
		normalized := strings.ToLower(strings.TrimSpace(roleName))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	if len(out) == 0 {
		out = []string{"user"}
	}

	return out
}

func hasRoleName(roleNames []string, roleName string) bool {
	target := strings.ToLower(strings.TrimSpace(roleName))
	if target == "" {
		return false
	}

	for _, candidate := range roleNames {
		if strings.ToLower(strings.TrimSpace(candidate)) == target {
			return true
		}
	}

	return false
}

func (h *Handler) resolveRoleIDsByName(roleNames []string) (map[string]struct{}, error) {
	desiredRoleIDs := make(map[string]struct{}, len(roleNames))

	for _, roleName := range roleNames {
		role, err := h.roles.GetRoleByName(roleName)
		if err != nil {
			return nil, err
		}
		if role == nil {
			return nil, fmt.Errorf("unknown role: %s", roleName)
		}
		desiredRoleIDs[role.ID] = struct{}{}
	}

	return desiredRoleIDs, nil
}

func (h *Handler) setUserRoles(userID string, roleNames []string) error {
	normalizedRoleNames := normalizeRoleNames(roleNames)
	desiredRoleIDs, err := h.resolveRoleIDsByName(normalizedRoleNames)
	if err != nil {
		return err
	}

	existingBindings, err := h.bindings.ListBindingsForUser(userID)
	if err != nil {
		return err
	}

	for _, binding := range existingBindings {
		if _, keep := desiredRoleIDs[binding.RoleID]; keep {
			delete(desiredRoleIDs, binding.RoleID)
			continue
		}
		if err := h.authorizer.RevokeRole(binding.ID); err != nil {
			return err
		}
	}

	for roleID := range desiredRoleIDs {
		if _, err := h.authorizer.AssignRole(userID, roleID); err != nil {
			return err
		}
	}

	return nil
}

// ═══════════════════════════════════════════════
// AUTH ENDPOINTS (public)
// ═══════════════════════════════════════════════

// Login authenticates a user and issues tokens.
func (h *Handler) Login(c *gin.Context) {
	var req authn.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.authn.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Resolve roles for the user
	roleNames, _ := h.authorizer.GetUserRoleNames(user.ID)

	// Create session
	session, err := h.authn.CreateSession(user.ID, c.ClientIP(), c.GetHeader("User-Agent"), h.issuer.AccessTokenTTL+time.Hour)
	if err != nil {
		log.Printf("⚠️  IAM: session creation failed: %v", err)
	}

	sessionID := ""
	if session != nil {
		sessionID = session.ID
	}

	pair, err := h.issuer.IssueTokenPair(user.ID, user.Email, user.DisplayName, "openid profile email roles", "", sessionID, roleNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token issuance failed"})
		return
	}

	c.JSON(http.StatusOK, authn.LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    pair.TokenType,
		ExpiresIn:    pair.ExpiresIn,
		ExpiresAt:    pair.ExpiresAt,
		User: authn.UserInfo{
			ID:            user.ID,
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			Roles:         roleNames,
			EmailVerified: user.EmailVerified,
		},
	})
}

// RefreshToken issues new tokens from a valid refresh token.
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	claims, err := h.issuer.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	// Look up user to ensure still active
	user, err := h.users.GetByID(claims.Subject)
	if err != nil || user == nil || !user.Active {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user account not found or disabled"})
		return
	}

	roleNames, _ := h.authorizer.GetUserRoleNames(user.ID)

	pair, err := h.issuer.IssueTokenPair(user.ID, user.Email, user.DisplayName, "openid profile email roles", "", "", roleNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token re-issuance failed"})
		return
	}

	c.JSON(http.StatusOK, pair)
}

// WhoAmI returns the currently authenticated user's info.
func (h *Handler) WhoAmI(c *gin.Context) {
	claims := iammw.GetClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":      claims.Sub,
		"email":        claims.Email,
		"display_name": claims.DisplayName,
		"roles":        claims.Roles,
	})
}

// ═══════════════════════════════════════════════
// USER MANAGEMENT (sysadmin only)
// ═══════════════════════════════════════════════

// ListUsers returns all IAM users.
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.users.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	for _, user := range users {
		if user == nil {
			continue
		}
		roleNames, _ := h.authorizer.GetUserRoleNames(user.ID)
		user.Roles = roleNames
	}

	c.JSON(http.StatusOK, gin.H{"users": users, "count": len(users)})
}

// GetUser returns a single user.
func (h *Handler) GetUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
		return
	}
	user, err := h.users.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	roleNames, _ := h.authorizer.GetUserRoleNames(user.ID)
	user.Roles = roleNames
	c.JSON(http.StatusOK, user)
}

// CreateUser registers a new IAM user.
func (h *Handler) CreateUser(c *gin.Context) {
	var req struct {
		Email         string   `json:"email" binding:"required,email"`
		Password      string   `json:"password" binding:"required,min=8"`
		DisplayName   string   `json:"display_name"`
		Active        *bool    `json:"active"`
		EmailVerified *bool    `json:"email_verified"`
		RoleNames     []string `json:"role_names"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	desiredRoleNames := normalizeRoleNames(req.RoleNames)
	if _, err := h.resolveRoleIDsByName(desiredRoleNames); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := identity.NormaliseEmail(req.Email)
	existing, _ := h.users.GetByEmail(email)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hash, err := identity.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &identity.User{
		ID:           identity.NewUserID(),
		Email:        email,
		PasswordHash: hash,
		DisplayName:  strings.TrimSpace(req.DisplayName),
		Active:       true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if req.Active != nil {
		user.Active = *req.Active
	}
	if req.EmailVerified != nil {
		user.EmailVerified = *req.EmailVerified
	}

	if err := h.users.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	if err := h.setUserRoles(user.ID, desiredRoleNames); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user created but role assignment failed: " + err.Error()})
		return
	}

	roleNames, _ := h.authorizer.GetUserRoleNames(user.ID)
	user.Roles = roleNames

	c.JSON(http.StatusCreated, user)
}

// UpdateUser modifies a user record.
func (h *Handler) UpdateUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
		return
	}

	user, err := h.users.GetByID(id)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req identity.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Email != nil {
		user.Email = identity.NormaliseEmail(*req.Email)
	}
	if req.Password != nil {
		hash, err := identity.HashPassword(*req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user.PasswordHash = hash
	}
	if req.DisplayName != nil {
		user.DisplayName = strings.TrimSpace(*req.DisplayName)
	}
	if req.Active != nil {
		user.Active = *req.Active
	}
	if req.EmailVerified != nil {
		user.EmailVerified = *req.EmailVerified
	}
	user.UpdatedAt = time.Now().UTC()

	if err := h.users.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// SetUserRoles replaces the role mapping for a user.
func (h *Handler) SetUserRoles(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
		return
	}

	user, err := h.users.GetByID(id)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req struct {
		RoleNames []string `json:"role_names" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	normalizedRoleNames := normalizeRoleNames(req.RoleNames)
	actorUserID := iammw.GetUserID(c)
	if actorUserID == user.ID && !hasRoleName(normalizedRoleNames, "sysadmin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot remove your own sysadmin role"})
		return
	}

	if err := h.setUserRoles(user.ID, normalizedRoleNames); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unknown role") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user roles"})
		return
	}

	roleNames, _ := h.authorizer.GetUserRoleNames(user.ID)
	c.JSON(http.StatusOK, gin.H{
		"user_id": user.ID,
		"roles":   roleNames,
		"count":   len(roleNames),
	})
}

// DeleteUser removes a user.
func (h *Handler) DeleteUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
		return
	}
	if err := h.users.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// ═══════════════════════════════════════════════
// OAUTH CLIENT MANAGEMENT (sysadmin only)
// ═══════════════════════════════════════════════

// RegisterClient creates a new OAuth2 client.
func (h *Handler) RegisterClient(c *gin.Context) {
	var req struct {
		Name         string   `json:"name" binding:"required"`
		RedirectURIs []string `json:"redirect_uris" binding:"required"`
		Scopes       []string `json:"scopes"`
		GrantTypes   []string `json:"grant_types"`
		Public       bool     `json:"public"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.GrantTypes) == 0 {
		req.GrantTypes = []string{"authorization_code", "refresh_token"}
	}
	if len(req.Scopes) == 0 {
		req.Scopes = []string{"openid", "profile", "email"}
	}

	secret := ""
	secretHash := ""
	if !req.Public {
		secret = oauth.GenerateClientSecret()
		var err error
		secretHash, err = identity.HashPassword(secret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "secret generation failed"})
			return
		}
	}

	client := &oauth.OAuthClient{
		ID:           oauth.GenerateClientID(),
		Secret:       secretHash,
		Name:         strings.TrimSpace(req.Name),
		RedirectURIs: req.RedirectURIs,
		Scopes:       req.Scopes,
		GrantTypes:   req.GrantTypes,
		Public:       req.Public,
		CreatedAt:    time.Now().UTC(),
		Active:       true,
	}

	if err := h.clients.CreateClient(client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create client"})
		return
	}

	// Return secret in plaintext only during creation
	resp := gin.H{
		"id":            client.ID,
		"name":          client.Name,
		"redirect_uris": client.RedirectURIs,
		"scopes":        client.Scopes,
		"grant_types":   client.GrantTypes,
		"public":        client.Public,
		"created_at":    client.CreatedAt,
	}
	if secret != "" {
		resp["client_secret"] = secret
		resp["warning"] = "Store the client_secret securely. It will not be shown again."
	}

	c.JSON(http.StatusCreated, resp)
}

// ListClients returns all OAuth clients.
func (h *Handler) ListClients(c *gin.Context) {
	clients, err := h.clients.ListClients()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list clients"})
		return
	}
	// Never expose secrets in list
	safe := make([]gin.H, 0, len(clients))
	for _, cl := range clients {
		safe = append(safe, gin.H{
			"id":            cl.ID,
			"name":          cl.Name,
			"redirect_uris": cl.RedirectURIs,
			"scopes":        cl.Scopes,
			"grant_types":   cl.GrantTypes,
			"public":        cl.Public,
			"active":        cl.Active,
			"created_at":    cl.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"clients": safe, "count": len(safe)})
}

// GetClient returns a single client (secret hidden).
func (h *Handler) GetClient(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	client, err := h.clients.GetClient(id)
	if err != nil || client == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":            client.ID,
		"name":          client.Name,
		"redirect_uris": client.RedirectURIs,
		"scopes":        client.Scopes,
		"grant_types":   client.GrantTypes,
		"public":        client.Public,
		"active":        client.Active,
		"created_at":    client.CreatedAt,
	})
}

// UpdateClient modifies a client's redirect URIs, scopes, or active status.
func (h *Handler) UpdateClient(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	client, err := h.clients.GetClient(id)
	if err != nil || client == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}

	var req struct {
		RedirectURIs *[]string `json:"redirect_uris"`
		Scopes       *[]string `json:"scopes"`
		Active       *bool     `json:"active"`
		Name         *string   `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RedirectURIs != nil {
		client.RedirectURIs = *req.RedirectURIs
	}
	if req.Scopes != nil {
		client.Scopes = *req.Scopes
	}
	if req.Active != nil {
		client.Active = *req.Active
	}
	if req.Name != nil {
		client.Name = strings.TrimSpace(*req.Name)
	}

	if err := h.clients.UpdateClient(client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "client updated", "id": client.ID})
}

// DeleteClient removes a client.
func (h *Handler) DeleteClient(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if err := h.clients.DeleteClient(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "client deleted"})
}

// ═══════════════════════════════════════════════
// ROLE MANAGEMENT (sysadmin only)
// ═══════════════════════════════════════════════

// CreateRole creates a new IAM role.
func (h *Handler) CreateRole(c *gin.Context) {
	var req authz.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, _ := h.roles.GetRoleByName(req.Name)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "role name already exists"})
		return
	}

	now := time.Now().UTC()
	role := &authz.Role{
		ID:          "role-" + strings.ToLower(strings.ReplaceAll(req.Name, " ", "-")) + "-" + now.Format("20060102150405"),
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		System:      false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.roles.CreateRole(role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create role"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

// ListRoles returns all IAM roles.
func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.roles.ListRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles, "count": len(roles)})
}

// GetRole returns a single role.
func (h *Handler) GetRole(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	role, err := h.roles.GetRole(id)
	if err != nil || role == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// UpdateRole modifies a role.
func (h *Handler) UpdateRole(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	role, err := h.roles.GetRole(id)
	if err != nil || role == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if role.System {
		c.JSON(http.StatusForbidden, gin.H{"error": "system roles cannot be modified"})
		return
	}

	var req struct {
		Description *string             `json:"description"`
		Permissions *[]authz.Permission `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Permissions != nil {
		role.Permissions = *req.Permissions
	}
	role.UpdatedAt = time.Now().UTC()

	if err := h.roles.UpdateRole(role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// DeleteRole removes a non-system role.
func (h *Handler) DeleteRole(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	role, err := h.roles.GetRole(id)
	if err != nil || role == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if role.System {
		c.JSON(http.StatusForbidden, gin.H{"error": "system roles cannot be deleted"})
		return
	}
	if err := h.roles.DeleteRole(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role deleted"})
}

// ═══════════════════════════════════════════════
// ROLE ASSIGNMENT (sysadmin only)
// ═══════════════════════════════════════════════

// AssignRole binds a user to a role.
func (h *Handler) AssignRole(c *gin.Context) {
	var req authz.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	binding, err := h.authorizer.AssignRole(req.UserID, req.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, binding)
}

// ListBindings returns role bindings, optionally filtered by user_id query param.
func (h *Handler) ListBindings(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("user_id"))

	var bindings []*authz.RoleBinding
	var err error
	if userID != "" {
		bindings, err = h.bindings.ListBindingsForUser(userID)
	} else {
		bindings, err = h.bindings.ListAllBindings()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bindings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bindings": bindings, "count": len(bindings)})
}

// RevokeBinding removes a role binding.
func (h *Handler) RevokeBinding(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if err := h.authorizer.RevokeRole(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "revoke failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role binding revoked"})
}

// ═══════════════════════════════════════════════
// TOKEN MANAGEMENT (sysadmin only)
// ═══════════════════════════════════════════════

// RevokeToken marks a token JTI as revoked.
func (h *Handler) RevokeToken(c *gin.Context) {
	var req struct {
		JTI string `json:"jti" binding:"required"`
		TTL int    `json:"ttl_seconds"` // how long to remember the revocation
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ttl := time.Duration(req.TTL) * time.Second
	if ttl <= 0 {
		ttl = time.Hour
	}

	if err := h.revokedStore.Revoke(req.JTI, ttl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "revocation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "token revoked", "jti": req.JTI})
}

// RevokeUserTokens revokes all refresh tokens for a user.
func (h *Handler) RevokeUserTokens(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
		return
	}

	if err := h.refreshRepo.RevokeAllForUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "revocation failed"})
		return
	}
	// Also revoke sessions
	_ = h.sessions.RevokeByUserID(userID)

	c.JSON(http.StatusOK, gin.H{"message": "all tokens revoked for user", "user_id": userID})
}

// ═══════════════════════════════════════════════
// OAUTH2 ENDPOINTS (Authorization Code + PKCE)
// ═══════════════════════════════════════════════

// Authorize handles GET /oauth/authorize (Authorization Code + PKCE).
// In a real deployment this would render a consent screen; for API-first usage
// it validates parameters and issues the code directly to authenticated users.
func (h *Handler) Authorize(c *gin.Context) {
	var req oauth.AuthorizeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorize request: " + err.Error()})
		return
	}

	if req.ResponseType != "code" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only response_type=code is supported"})
		return
	}
	if req.CodeChallengeMethod != "S256" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only S256 code_challenge_method is supported (PKCE required)"})
		return
	}

	client, err := h.clients.GetClient(req.ClientID)
	if err != nil || client == nil || !client.Active {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown or inactive client_id"})
		return
	}

	if err := oauth.ValidateRedirectURI(client.RedirectURIs, req.RedirectURI); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scopes := oauth.ParseScopes(req.Scope)
	if len(scopes) > 0 {
		if err := oauth.ValidateScopes(client.Scopes, scopes); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// The user must be authenticated (via IAM JWT)
	userID := iammw.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user must be authenticated to authorize"})
		return
	}

	code, err := oauth.GenerateAuthorizationCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "code generation failed"})
		return
	}

	codeRecord := &oauth.AuthorizationCode{
		Code:                code,
		ClientID:            req.ClientID,
		UserID:              userID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().UTC().Add(5 * time.Minute),
	}

	if h.codeRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "code storage not initialised"})
		return
	}
	if err := h.codeRepo.StoreCode(codeRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store authorization code"})
		return
	}

	// Redirect back with code
	c.JSON(http.StatusOK, gin.H{
		"code":         code,
		"state":        req.State,
		"redirect_uri": req.RedirectURI,
	})
}

// Token handles POST /oauth/token (code exchange + refresh).
func (h *Handler) Token(c *gin.Context) {
	var req oauth.TokenRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token request"})
		return
	}

	switch req.GrantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(c, &req)
	case "refresh_token":
		h.handleRefreshTokenGrant(c, &req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported grant_type"})
	}
}

func (h *Handler) handleAuthorizationCodeGrant(c *gin.Context, req *oauth.TokenRequest) {
	if h.codeRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "code storage not initialised"})
		return
	}

	if req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	codeRecord, err := h.codeRepo.GetCode(req.Code)
	if err != nil || codeRecord == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired authorization code"})
		return
	}

	// Invalidate the code immediately (single-use)
	_ = h.codeRepo.InvalidateCode(req.Code)

	if codeRecord.Used {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code already used"})
		return
	}

	if time.Now().UTC().After(codeRecord.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code expired"})
		return
	}

	if codeRecord.ClientID != req.ClientID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id mismatch"})
		return
	}

	if codeRecord.RedirectURI != req.RedirectURI {
		c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri mismatch"})
		return
	}

	// Verify PKCE
	if err := oauth.VerifyPKCE(codeRecord.CodeChallenge, req.CodeVerifier, codeRecord.CodeChallengeMethod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate client secret (for confidential clients)
	client, err := h.clients.GetClient(req.ClientID)
	if err != nil || client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown client"})
		return
	}
	if !client.Public && client.Secret != "" {
		if !identity.CheckPassword(req.ClientSecret, client.Secret) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client_secret"})
			return
		}
	}

	// Look up user
	user, err := h.users.GetByID(codeRecord.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	roleNames, _ := h.authorizer.GetUserRoleNames(user.ID)

	pair, err := h.issuer.IssueTokenPair(user.ID, user.Email, user.DisplayName, codeRecord.Scope, req.ClientID, "", roleNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token issuance failed"})
		return
	}

	c.JSON(http.StatusOK, pair)
}

func (h *Handler) handleRefreshTokenGrant(c *gin.Context, req *oauth.TokenRequest) {
	if req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	claims, err := h.issuer.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	user, err := h.users.GetByID(claims.Subject)
	if err != nil || user == nil || !user.Active {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found or disabled"})
		return
	}

	roleNames, _ := h.authorizer.GetUserRoleNames(user.ID)

	scope := req.Scope
	if scope == "" {
		scope = "openid profile email roles"
	}

	pair, err := h.issuer.IssueTokenPair(user.ID, user.Email, user.DisplayName, scope, req.ClientID, "", roleNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token re-issuance failed"})
		return
	}

	c.JSON(http.StatusOK, pair)
}

// ═══════════════════════════════════════════════
// OIDC DISCOVERY
// ═══════════════════════════════════════════════

// OpenIDConfiguration returns the OIDC discovery document.
func (h *Handler) OpenIDConfiguration(c *gin.Context) {
	c.JSON(http.StatusOK, h.issuer.OpenIDConfiguration())
}

// JWKS returns the JSON Web Key Set.
func (h *Handler) JWKS(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=3600")
	c.JSON(http.StatusOK, h.issuer.JWKS())
}
