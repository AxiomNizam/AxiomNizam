package handlers

import (
	"net/http"

	"example.com/axiomnizam/internal/gatekeeper/contracts"
	"example.com/axiomnizam/internal/gatekeeper/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HTTPHandler provides REST endpoints for 2FA operations.
type HTTPHandler struct {
	enrollmentSvc contracts.EnrollmentService
	challengeSvc  contracts.ChallengeService
	factorSvc     contracts.FactorService
	policyService contracts.PolicyService
	riskSvc       contracts.RiskService
	deviceSvc     contracts.TrustedDeviceService
	backupSvc     contracts.BackupCodeService
}

// NewHTTPHandler creates a new HTTP handler.
func NewHTTPHandler(
	es contracts.EnrollmentService,
	cs contracts.ChallengeService,
	fs contracts.FactorService,
	ps contracts.PolicyService,
	rs contracts.RiskService,
	ds contracts.TrustedDeviceService,
	bs contracts.BackupCodeService,
) *HTTPHandler {
	return &HTTPHandler{
		enrollmentSvc: es,
		challengeSvc:  cs,
		factorSvc:     fs,
		policyService: ps,
		riskSvc:       rs,
		deviceSvc:     ds,
		backupSvc:     bs,
	}
}

// RegisterRoutes registers all HTTP endpoints under /api/v1/mfa.
func (h *HTTPHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/mfa")

	// Enrollment endpoints
	api.POST("/enroll", h.EnrollFactor)
	api.POST("/activate", h.ActivateFactor)
	api.POST("/disable/:factorID", h.DisableFactor)

	// Factor endpoints
	api.GET("/factors/:userID", h.ListFactors)
	api.GET("/factor/:factorID", h.GetFactor)

	// Challenge endpoints
	api.POST("/challenge/begin", h.BeginChallenge)
	api.POST("/challenge/verify", h.VerifyChallenge)

	// Policy endpoints
	api.GET("/policy/:userID", h.EvaluatePolicy)

	// Risk endpoints
	api.POST("/risk/score", h.ScoreRisk)

	// Trusted device endpoints
	api.POST("/trust-device", h.TrustDevice)
	api.GET("/trust-device/list/:userID", h.ListTrustedDevices)
	api.DELETE("/trust-device/:deviceID", h.RevokeTrustedDevice)
}

// EnrollFactor enrolls a new factor for a user.
func (h *HTTPHandler) EnrollFactor(c *gin.Context) {
	var req struct {
		UserID     uuid.UUID         `json:"user_id" binding:"required"`
		FactorType models.FactorType `json:"factor_type" binding:"required"`
		Email      string            `json:"email"`
		Phone      string            `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secret, err := h.enrollmentSvc.SetupFactor(c.Request.Context(), req.UserID, req.FactorType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"secret": secret,
	})
}

// ActivateFactor completes factor activation.
func (h *HTTPHandler) ActivateFactor(c *gin.Context) {
	var req struct {
		FactorID uuid.UUID `json:"factor_id" binding:"required"`
		Code     string    `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	codes, err := h.enrollmentSvc.ActivateFactor(c.Request.Context(), req.FactorID, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"backup_codes": codes,
	})
}

// DisableFactor disables a factor.
func (h *HTTPHandler) DisableFactor(c *gin.Context) {
	factorID := uuid.MustParse(c.Param("factorID"))

	if err := h.enrollmentSvc.DisableFactor(c.Request.Context(), factorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Factor disabled"})
}

// ListFactors lists all factors for a user.
func (h *HTTPHandler) ListFactors(c *gin.Context) {
	userID := uuid.MustParse(c.Param("userID"))

	factors, err := h.factorSvc.ListFactors(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Strip sensitive fields before returning
	for _, f := range factors {
		f.Spec.EncryptedSecret = nil
	}

	c.JSON(http.StatusOK, gin.H{"factors": factors})
}

// GetFactor retrieves a single factor.
func (h *HTTPHandler) GetFactor(c *gin.Context) {
	factorID := uuid.MustParse(c.Param("factorID"))

	factor, err := h.factorSvc.GetFactor(c.Request.Context(), factorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Strip sensitive fields before returning
	if factor != nil {
		factor.Spec.EncryptedSecret = nil
	}

	c.JSON(http.StatusOK, factor)
}

// BeginChallenge starts a new MFA challenge.
func (h *HTTPHandler) BeginChallenge(c *gin.Context) {
	var req struct {
		UserID   uuid.UUID `json:"user_id" binding:"required"`
		FactorID uuid.UUID `json:"factor_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	challengeID, err := h.challengeSvc.BeginChallenge(c.Request.Context(), req.UserID, req.FactorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"challenge_id": challengeID})
}

// VerifyChallenge verifies a challenge response.
func (h *HTTPHandler) VerifyChallenge(c *gin.Context) {
	var req struct {
		ChallengeID string `json:"challenge_id" binding:"required"`
		Code        string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	verified, err := h.challengeSvc.VerifyChallenge(c.Request.Context(), req.ChallengeID, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"verified": verified})
}

// EvaluatePolicy evaluates if MFA is required for a user.
func (h *HTTPHandler) EvaluatePolicy(c *gin.Context) {
	userID := uuid.MustParse(c.Param("userID"))

	requiresMFA, factors, err := h.policyService.EvaluatePolicy(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"requires_mfa": requiresMFA,
		"factors":      factors,
	})
}

// ScoreRisk scores the risk of an authentication request.
func (h *HTTPHandler) ScoreRisk(c *gin.Context) {
	var req struct {
		UserID    uuid.UUID `json:"user_id" binding:"required"`
		IPAddress string    `json:"ip_address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score, err := h.riskSvc.ScoreAuthentication(c.Request.Context(), req.UserID, req.IPAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"risk_score": score})
}

// TrustDevice registers a trusted device.
func (h *HTTPHandler) TrustDevice(c *gin.Context) {
	var req struct {
		UserID      uuid.UUID `json:"user_id" binding:"required"`
		Fingerprint string    `json:"fingerprint" binding:"required"`
		UserAgent   string    `json:"user_agent" binding:"required"`
		IPAddress   string    `json:"ip_address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.deviceSvc.TrustDevice(c.Request.Context(), req.UserID, req.Fingerprint, req.UserAgent, req.IPAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// ListTrustedDevices lists trusted devices for a user.
func (h *HTTPHandler) ListTrustedDevices(c *gin.Context) {
	// Placeholder implementation
	c.JSON(http.StatusOK, gin.H{"devices": []interface{}{}})
}

// RevokeTrustedDevice revokes a trusted device.
func (h *HTTPHandler) RevokeTrustedDevice(c *gin.Context) {
	deviceID := uuid.MustParse(c.Param("deviceID"))

	if err := h.deviceSvc.RevokeTrustedDevice(c.Request.Context(), deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device revoked"})
}
