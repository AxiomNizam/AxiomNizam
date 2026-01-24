package services

import (
	"context"
	"fmt"
	"log"

	"example.com/axiomnizam/internal/models"
	"example.com/axiomnizam/internal/repositories"
	"example.com/axiomnizam/internal/utils"
)

// AuthService defines the interface for authentication business logic
type AuthService interface {
	// Login authenticates a user with username and password
	Login(ctx context.Context, username string, password string) (*models.User, string, error)

	// Register registers a new user
	Register(ctx context.Context, user *models.User, password string) (*models.User, error)

	// ValidateToken validates an authentication token
	ValidateToken(ctx context.Context, token string) (bool, error)

	// RefreshToken refreshes an authentication token
	RefreshToken(ctx context.Context, token string) (string, error)

	// Logout logs out a user
	Logout(ctx context.Context, token string) error

	// Health checks if the service is healthy
	Health() error
}

// authService implements AuthService interface
type authService struct {
	*BaseService
	userRepo repositories.UserRepository
	logger   *log.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repositories.UserRepository, validator *utils.InputValidator, sqlProtection *utils.SQLInjectionProtection) AuthService {
	return &authService{
		BaseService: NewBaseService(validator, sqlProtection),
		userRepo:    userRepo,
		logger:      log.New(log.Writer(), "[AUTH_SERVICE] ", log.LstdFlags),
	}
}

// Login authenticates a user with username and password
// Returns the user and a JWT token if successful
func (s *authService) Login(ctx context.Context, username string, password string) (*models.User, string, error) {
	if username == "" || password == "" {
		s.LogError("Login", fmt.Errorf("username or password is empty"))
		return nil, "", ErrInvalidInput
	}

	// Validate input
	batch := s.GetValidator().NewValidationBatch().
		AddStringValidation("username", username, utils.WithMinLength(3)).
		AddStringValidation("password", password, utils.WithMinLength(1))

	if batch.HasErrors() {
		s.LogError("Login validation", fmt.Errorf(batch.Error()))
		return nil, "", ErrValidationFailed
	}

	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if err == repositories.ErrNotFound {
			s.logger.Printf("LOGIN_ATTEMPT_FAILED: username=%s, reason=not_found", username)
			// Don't reveal if user exists
			return nil, "", ErrUnauthorized
		}
		s.LogError("GetByUsername", err)
		return nil, "", ErrInternalServer
	}

	// TODO: Verify password hash
	// This should use bcrypt or similar to compare hashed passwords
	// if !VerifyPasswordHash(user.PasswordHash, password) {
	//     s.logger.Printf("LOGIN_ATTEMPT_FAILED: username=%s, reason=invalid_password", username)
	//     return nil, "", ErrUnauthorized
	// }

	// Generate JWT token
	// TODO: Implement JWT token generation
	token := "jwt-token-placeholder"

	s.logger.Printf("LOGIN_SUCCESS: username=%s, user_id=%s", username, user.ID)
	return user, token, nil
}

// Register registers a new user with validation
func (s *authService) Register(ctx context.Context, user *models.User, password string) (*models.User, error) {
	if user == nil {
		return nil, ErrInvalidInput
	}

	if user.Email == "" || user.Username == "" {
		s.LogError("Register", fmt.Errorf("email or username is empty"))
		return nil, ErrInvalidInput
	}

	// Validate password strength
	if !utils.IsValidPassword(password) {
		s.LogError("Register", fmt.Errorf("weak password for user: %s", user.Email))
		return nil, ErrValidationFailed
	}

	// Validate input
	batch := s.GetValidator().NewValidationBatch().
		AddEmailValidation("email", user.Email).
		AddStringValidation("username", user.Username, utils.WithMinLength(3), utils.WithMaxLength(20)).
		AddStringValidation("password", password, utils.WithMinLength(8))

	if batch.HasErrors() {
		s.LogError("Register validation", fmt.Errorf(batch.Error()))
		return nil, ErrValidationFailed
	}

	// Check if email already exists
	existsByEmail, err := s.userRepo.ExistsByEmail(ctx, user.Email)
	if err != nil {
		s.LogError("Register ExistsByEmail", err)
		return nil, ErrInternalServer
	}
	if existsByEmail {
		s.logger.Printf("REGISTER_FAILED: email=%s, reason=already_exists", user.Email)
		return nil, ErrDuplicateEntry
	}

	// Check if username already exists
	existsByUsername, err := s.userRepo.ExistsByUsername(ctx, user.Username)
	if err != nil {
		s.LogError("Register ExistsByUsername", err)
		return nil, ErrInternalServer
	}
	if existsByUsername {
		s.logger.Printf("REGISTER_FAILED: username=%s, reason=already_exists", user.Username)
		return nil, ErrDuplicateEntry
	}

	// TODO: Hash password using bcrypt
	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// if err != nil {
	//     s.LogError("Register bcrypt", err)
	//     return nil, ErrInternalServer
	// }
	// user.PasswordHash = string(hashedPassword)

	// Create user
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.LogError("Register Create", err)
		return nil, ErrInternalServer
	}

	s.logger.Printf("REGISTER_SUCCESS: email=%s, username=%s, user_id=%s", user.Email, user.Username, user.ID)
	return user, nil
}

// ValidateToken validates an authentication token
func (s *authService) ValidateToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, ErrInvalidInput
	}

	// TODO: Implement JWT token validation
	// This should verify the token signature and expiration
	// For now, just check if token is not empty
	return token != "", nil
}

// RefreshToken refreshes an authentication token
func (s *authService) RefreshToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", ErrInvalidInput
	}

	// TODO: Implement JWT token refresh
	// This should verify the current token and return a new one
	return "new-jwt-token", nil
}

// Logout logs out a user
func (s *authService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidInput
	}

	// TODO: Implement logout logic
	// This might involve blacklisting the token or removing it from cache
	return nil
}

// Health checks if the service is healthy
func (s *authService) Health() error {
	// Try to get a user to check if repository is accessible
	ctx := context.Background()
	_, err := s.userRepo.GetByUsername(ctx, "test-user")
	// We don't care if the user exists, just if we can query
	if err != nil && err != repositories.ErrNotFound {
		return err
	}
	return nil
}
