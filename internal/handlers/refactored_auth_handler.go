package handlers

import (
	"context"
	"net/http"

	"example.com/axiomnizam/internal/models"
	"example.com/axiomnizam/internal/services"
	"github.com/gin-gonic/gin"
)

// RefactoredAuthHandler uses the AuthService for business logic
type RefactoredAuthHandler struct {
	authService services.AuthService
}

// NewRefactoredAuthHandler creates a new auth handler with service injection
func NewRefactoredAuthHandler(authService services.AuthService) *RefactoredAuthHandler {
	return &RefactoredAuthHandler{
		authService: authService,
	}
}

// LoginResponse represents the login response
type LoginResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles POST /auth/login - Now uses service layer
func (h *RefactoredAuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid request: " + err.Error(),
		})
		return
	}

	ctx := context.Background()
	user, token, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		switch err {
		case services.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, models.Response{
				Status: "error",
				Error:  "Invalid username or password",
			})
		case services.ErrUnauthorized:
			c.JSON(http.StatusUnauthorized, models.Response{
				Status: "error",
				Error:  "Invalid username or password",
			})
		case services.ErrValidationFailed:
			c.JSON(http.StatusBadRequest, models.Response{
				Status: "error",
				Error:  "Validation failed",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  "Login failed",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status: "success",
		Data: LoginResponse{
			User:  user,
			Token: token,
		},
	})
}

// Register handles POST /auth/register - Now uses service layer
func (h *RefactoredAuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid request: " + err.Error(),
		})
		return
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
	}

	ctx := context.Background()
	createdUser, err := h.authService.Register(ctx, user, req.Password)
	if err != nil {
		switch err {
		case services.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, models.Response{
				Status: "error",
				Error:  "Invalid input",
			})
		case services.ErrDuplicateEntry:
			c.JSON(http.StatusConflict, models.Response{
				Status: "error",
				Error:  "Email or username already exists",
			})
		case services.ErrValidationFailed:
			c.JSON(http.StatusBadRequest, models.Response{
				Status: "error",
				Error:  "Invalid email or weak password",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  "Registration failed",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Status: "success",
		Data:   createdUser,
	})
}

// ValidateToken handles GET /auth/validate
func (h *RefactoredAuthHandler) ValidateToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Missing authorization token",
		})
		return
	}

	ctx := context.Background()
	valid, err := h.authService.ValidateToken(ctx, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Token validation failed",
		})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, models.Response{
			Status: "error",
			Error:  "Invalid token",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status: "success",
		Data: gin.H{
			"valid": true,
		},
	})
}

// RefreshToken handles POST /auth/refresh
func (h *RefactoredAuthHandler) RefreshToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Missing authorization token",
		})
		return
	}

	ctx := context.Background()
	newToken, err := h.authService.RefreshToken(ctx, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Token refresh failed",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status: "success",
		Data: gin.H{
			"token": newToken,
		},
	})
}

// Logout handles POST /auth/logout
func (h *RefactoredAuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Missing authorization token",
		})
		return
	}

	ctx := context.Background()
	if err := h.authService.Logout(ctx, token); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Logout failed",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  "success",
		Message: "Logged out successfully",
	})
}
