package handlers

import (
	"fmt"
	"net/http"

	"example.com/axiomnizam/internal/models"
	"example.com/axiomnizam/internal/services"
	"github.com/gin-gonic/gin"
)

// RefactoredUserHandler uses the UserService for business logic
type RefactoredUserHandler struct {
	userService services.UserService
}

// NewRefactoredUserHandler creates a new user handler with service injection
func NewRefactoredUserHandler(userService services.UserService) *RefactoredUserHandler {
	return &RefactoredUserHandler{
		userService: userService,
	}
}

// CreateUser handles POST /users - Now uses service layer
func (h *RefactoredUserHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid request: " + err.Error(),
		})
		return
	}

	// Use service to create user with validation and business logic
	ctx := c.Request.Context()
	createdUser, err := h.userService.CreateUser(ctx, &user)
	if err != nil {
		// Handle service errors appropriately
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
				Error:  "Validation failed",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  "Failed to create user",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Status: "success",
		Data:   createdUser,
	})
}

// GetUser handles GET /users/:id - Now uses service layer
func (h *RefactoredUserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	user, err := h.userService.GetUserByID(ctx, id)
	if err != nil {
		switch err {
		case services.ErrNotFound:
			c.JSON(http.StatusNotFound, models.Response{
				Status: "error",
				Error:  "User not found",
			})
		case services.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, models.Response{
				Status: "error",
				Error:  "Invalid user ID",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  "Failed to get user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status: "success",
		Data:   user,
	})
}

// GetUserByEmail handles GET /users/email/:email - Now uses service layer
func (h *RefactoredUserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")

	ctx := c.Request.Context()
	user, err := h.userService.GetUserByEmail(ctx, email)
	if err != nil {
		switch err {
		case services.ErrNotFound:
			c.JSON(http.StatusNotFound, models.Response{
				Status: "error",
				Error:  "User not found",
			})
		case services.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, models.Response{
				Status: "error",
				Error:  "Invalid email format",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  "Failed to get user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status: "success",
		Data:   user,
	})
}

// ListUsers handles GET /users - Now uses service layer
func (h *RefactoredUserHandler) ListUsers(c *gin.Context) {
	// Get pagination parameters
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		// Parse limit from query, with validation
		if _, err := c.GetQuery("limit"); err {
			// Implement limit validation here
		}
	}

	ctx := c.Request.Context()
	users, err := h.userService.ListUsers(ctx, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to list users",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status: "success",
		Data:   users,
	})
}

// UpdateUser handles PUT /users/:id - Now uses service layer
func (h *RefactoredUserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid request: " + err.Error(),
		})
		return
	}

	var userID uint
	if _, err := fmt.Sscanf(id, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid user ID format",
		})
		return
	}
	user.ID = userID

	ctx := c.Request.Context()
	updatedUser, err := h.userService.UpdateUser(ctx, &user)
	if err != nil {
		switch err {
		case services.ErrNotFound:
			c.JSON(http.StatusNotFound, models.Response{
				Status: "error",
				Error:  "User not found",
			})
		case services.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, models.Response{
				Status: "error",
				Error:  "Invalid input",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  "Failed to update user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status: "success",
		Data:   updatedUser,
	})
}

// DeleteUser handles DELETE /users/:id - Now uses service layer
func (h *RefactoredUserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	err := h.userService.DeleteUser(ctx, id)
	if err != nil {
		switch err {
		case services.ErrNotFound:
			c.JSON(http.StatusNotFound, models.Response{
				Status: "error",
				Error:  "User not found",
			})
		case services.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, models.Response{
				Status: "error",
				Error:  "Invalid user ID",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  "Failed to delete user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  "success",
		Message: "User deleted successfully",
	})
}
