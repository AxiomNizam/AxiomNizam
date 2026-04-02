package iam

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"example.com/axiomnizam/internal/iam/admin"
	"example.com/axiomnizam/internal/iam/authn"
	"example.com/axiomnizam/internal/iam/authz"
	"example.com/axiomnizam/internal/iam/identity"
	iammw "example.com/axiomnizam/internal/iam/middleware"
	"example.com/axiomnizam/internal/iam/storage"
	"example.com/axiomnizam/internal/iam/token"
	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

// System holds the fully initialised IAM system and exposes the router registration.
type System struct {
	Issuer        *token.Issuer
	Authorizer    *authz.Authorizer
	Authenticator *authn.Authenticator
	AdminHandler  *admin.Handler

	// Repositories (exported for external callers that need direct access)
	Users        *storage.PostgresUserRepository
	Clients      *storage.EtcdClientRepository
	Roles        *storage.EtcdRoleRepository
	Bindings     *storage.EtcdRoleBindingRepository
	Sessions     *storage.EtcdSessionRepository
	RefreshRepo  *storage.EtcdRefreshTokenRepository
	CodeRepo     *storage.EtcdCodeRepository
	RevokedStore *storage.EtcdRevokedTokenStore
}

// Config specifies IAM startup options.
type Config struct {
	// IssuerURL is the public URL of this IAM instance (e.g. https://api.example.com).
	IssuerURL string

	// SysadminEmail is the bootstrap sysadmin account email.
	SysadminEmail string
	// SysadminPassword is the bootstrap sysadmin password.
	SysadminPassword string

	// AccessTokenTTL overrides the default 15-minute access token lifetime.
	AccessTokenTTL time.Duration
	// RefreshTokenTTL overrides the default 7-day refresh token lifetime.
	RefreshTokenTTL time.Duration
}

// NewSystem initialises the complete IAM system.
// Requires a PostgreSQL connection (for users) and an etcd client (for everything else).
func NewSystem(pg *gorm.DB, etcd *clientv3.Client, cfg Config) (*System, error) {
	if pg == nil {
		return nil, errors.New("IAM requires a PostgreSQL connection")
	}
	if etcd == nil {
		return nil, errors.New("IAM requires an etcd connection")
	}

	// Token issuer
	issuerURL := cfg.IssuerURL
	if issuerURL == "" {
		issuerURL = os.Getenv("IAM_ISSUER_URL")
	}
	if issuerURL == "" {
		issuerURL = "http://localhost:8080"
	}

	issuer, err := token.NewIssuer(issuerURL)
	if err != nil {
		return nil, fmt.Errorf("IAM token issuer: %w", err)
	}
	if cfg.AccessTokenTTL > 0 {
		issuer.AccessTokenTTL = cfg.AccessTokenTTL
	}
	if cfg.RefreshTokenTTL > 0 {
		issuer.RefreshTokenTTL = cfg.RefreshTokenTTL
	}

	// Repositories
	userRepo, err := storage.NewPostgresUserRepository(pg)
	if err != nil {
		return nil, fmt.Errorf("IAM user repository: %w", err)
	}

	clientRepo := storage.NewEtcdClientRepository(etcd)
	roleRepo := storage.NewEtcdRoleRepository(etcd)
	bindingRepo := storage.NewEtcdRoleBindingRepository(etcd)
	sessionRepo := storage.NewEtcdSessionRepository(etcd)
	refreshRepo := storage.NewEtcdRefreshTokenRepository(etcd)
	codeRepo := storage.NewEtcdCodeRepository(etcd)
	revokedStore := storage.NewEtcdRevokedTokenStore(etcd)

	// Seed system roles
	if err := storage.SeedSystemRoles(roleRepo); err != nil {
		log.Printf("⚠️  IAM: role seeding error (non-fatal): %v", err)
	}

	// Authorizer
	authorizer := authz.NewAuthorizer(roleRepo, bindingRepo)

	// Authenticator
	authenticator := authn.NewAuthenticator(userRepo, sessionRepo)

	// Bootstrap sysadmin
	sysadminEmail := cfg.SysadminEmail
	if sysadminEmail == "" {
		sysadminEmail = os.Getenv("IAM_SYSADMIN_EMAIL")
	}
	sysadminPassword := cfg.SysadminPassword
	if sysadminPassword == "" {
		sysadminPassword = os.Getenv("IAM_SYSADMIN_PASSWORD")
	}

	if sysadminEmail != "" && sysadminPassword != "" {
		if err := bootstrapSysadmin(userRepo, roleRepo, authorizer, sysadminEmail, sysadminPassword); err != nil {
			log.Printf("⚠️  IAM: sysadmin bootstrap error (non-fatal): %v", err)
		}
	} else {
		log.Println("⚠️  IAM: No sysadmin credentials configured. Set IAM_SYSADMIN_EMAIL and IAM_SYSADMIN_PASSWORD.")
	}

	adminHandler := admin.NewHandler(
		userRepo, clientRepo, roleRepo, bindingRepo,
		sessionRepo, refreshRepo, revokedStore, codeRepo,
		authorizer, issuer, authenticator,
	)

	return &System{
		Issuer:        issuer,
		Authorizer:    authorizer,
		Authenticator: authenticator,
		AdminHandler:  adminHandler,
		Users:         userRepo,
		Clients:       clientRepo,
		Roles:         roleRepo,
		Bindings:      bindingRepo,
		Sessions:      sessionRepo,
		RefreshRepo:   refreshRepo,
		CodeRepo:      codeRepo,
		RevokedStore:  revokedStore,
	}, nil
}

// RegisterRoutes mounts all IAM endpoints on the provided Gin router.
func (s *System) RegisterRoutes(router *gin.Engine) {
	h := s.AdminHandler

	// ── OIDC Discovery (public) ──
	router.GET("/.well-known/openid-configuration", h.OpenIDConfiguration)
	router.GET("/.well-known/jwks.json", h.JWKS)
	router.GET("/realms/:realm/.well-known/openid-configuration", h.OpenIDConfigurationRealm)
	router.POST("/realms/:realm/protocol/openid-connect/token", h.RealmToken)
	router.GET("/realms/:realm/protocol/openid-connect/certs", h.RealmJWKS)

	// ── Authentication (public) ──
	router.POST("/iam/auth/login", h.Login)
	router.POST("/iam/auth/refresh", h.RefreshToken)

	// IAM JWT middleware for all protected IAM routes
	iamAuth := iammw.JWTAuth(s.Issuer, s.RevokedStore)
	sysadminOnly := iammw.RequireSysadmin()

	// ── Self-service (authenticated) ──
	router.GET("/iam/auth/whoami", iamAuth, h.WhoAmI)

	// ── OAuth2 Endpoints (authenticated user initiates, public token endpoint) ──
	router.GET("/oauth/authorize", iamAuth, h.Authorize)
	router.GET("/realms/:realm/protocol/openid-connect/auth", iamAuth, h.RealmAuthorize)
	router.POST("/oauth/token", h.Token)

	// ── Sysadmin API (master-realm admin only) ──
	adminAPI := router.Group("/iam/admin", iamAuth, sysadminOnly)
	{
		// Users
		adminAPI.GET("/users", h.ListUsers)
		adminAPI.GET("/users/:id", h.GetUser)
		adminAPI.POST("/users", h.CreateUser)
		adminAPI.PUT("/users/:id", h.UpdateUser)
		adminAPI.PUT("/users/:id/roles", h.SetUserRoles)
		adminAPI.DELETE("/users/:id", h.DeleteUser)

		// OAuth Clients
		adminAPI.GET("/clients", h.ListClients)
		adminAPI.GET("/clients/:id", h.GetClient)
		adminAPI.POST("/clients", h.RegisterClient)
		adminAPI.PUT("/clients/:id", h.UpdateClient)
		adminAPI.DELETE("/clients/:id", h.DeleteClient)
		adminAPI.GET("/service-access-info", h.ServiceAccessInfo)

		// Roles
		adminAPI.GET("/roles", h.ListRoles)
		adminAPI.GET("/roles/:id", h.GetRole)
		adminAPI.POST("/roles", h.CreateRole)
		adminAPI.PUT("/roles/:id", h.UpdateRole)
		adminAPI.DELETE("/roles/:id", h.DeleteRole)

		// Role Assignments
		adminAPI.POST("/role-bindings", h.AssignRole)
		adminAPI.GET("/role-bindings", h.ListBindings)
		adminAPI.DELETE("/role-bindings/:id", h.RevokeBinding)

		// Token Management
		adminAPI.POST("/tokens/revoke", h.RevokeToken)
		adminAPI.POST("/users/:id/revoke-tokens", h.RevokeUserTokens)
	}

	log.Println("✅ IAM: all routes registered")
}

// bootstrapSysadmin ensures the master-realm admin account exists.
func bootstrapSysadmin(
	userRepo *storage.PostgresUserRepository,
	roleRepo *storage.EtcdRoleRepository,
	authorizer *authz.Authorizer,
	email, password string,
) error {
	email = identity.NormaliseEmail(email)

	existing, _ := userRepo.GetByEmail(email)
	if existing != nil {
		log.Printf("✅ IAM: sysadmin account already exists (%s)", email)
		// Ensure sysadmin role is bound
		sysRole, _ := roleRepo.GetRoleByName("sysadmin")
		if sysRole != nil {
			_, _ = authorizer.AssignRole(existing.ID, sysRole.ID)
		}
		return nil
	}

	hash, err := identity.HashPassword(password)
	if err != nil {
		return fmt.Errorf("password hash: %w", err)
	}

	user := &identity.User{
		ID:            identity.NewUserID(),
		Email:         email,
		PasswordHash:  hash,
		DisplayName:   "System Administrator",
		Active:        true,
		EmailVerified: true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := userRepo.Create(user); err != nil {
		return fmt.Errorf("creating sysadmin: %w", err)
	}

	sysRole, _ := roleRepo.GetRoleByName("sysadmin")
	if sysRole != nil {
		if _, err := authorizer.AssignRole(user.ID, sysRole.ID); err != nil {
			return fmt.Errorf("assigning sysadmin role: %w", err)
		}
	}

	log.Printf("✅ IAM: sysadmin account bootstrapped (%s)", email)
	return nil
}
