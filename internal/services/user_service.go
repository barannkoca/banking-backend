package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/barannkoca/banking-backend/internal/interfaces"
	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserService implements the UserService interface
type UserService struct {
	userRepo     interfaces.UserRepository
	auditService interfaces.AuditService
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo interfaces.UserRepository, auditService interfaces.AuditService) *UserService {
	return &UserService{
		userRepo:     userRepo,
		auditService: auditService,
	}
}

// CreateUser creates a new user with password hashing
func (us *UserService) CreateUser(ctx context.Context, req *models.UserCreateRequest) (*models.User, error) {
	// Validate request
	if err := us.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if email is available
	emailExists, err := us.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("email availability check failed: %w", err)
	}
	if emailExists {
		return nil, fmt.Errorf("email already exists")
	}

	// Check if username is available
	usernameExists, err := us.userRepo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("username availability check failed: %w", err)
	}
	if usernameExists {
		return nil, fmt.Errorf("username already exists")
	}

	// Determine role
	role := models.RoleCustomer
	if req.Role != "" {
		switch strings.ToLower(req.Role) {
		case "admin":
			role = models.RoleAdmin
		case "teller":
			role = models.RoleTeller
		case "customer":
			role = models.RoleCustomer
		default:
			return nil, fmt.Errorf("invalid role: %s", req.Role)
		}
	}

	// Hash password
	hashedPassword, err := us.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Username:     strings.TrimSpace(req.Username),
		Email:        strings.TrimSpace(strings.ToLower(req.Email)),
		PasswordHash: hashedPassword,
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Validate user data
	if err := us.ValidateUserData(ctx, user); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Save to database
	if err := us.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Log audit trail
	if us.auditService != nil {
		us.auditService.LogUserActivity(ctx, user.ID, "USER_CREATED", "user", user.ID.String(),
			fmt.Sprintf("User created with role: %s", user.Role))
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (us *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := us.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// GetUserByUsername retrieves a user by username
func (us *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := us.userRepo.GetByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (us *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := us.userRepo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// UpdateUser updates user information
func (us *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	// Validate user data
	if err := us.ValidateUserData(ctx, user); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Check if user exists
	existingUser, err := us.GetUserByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Save to database
	if err := us.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Log audit trail
	if us.auditService != nil {
		us.auditService.LogUserActivity(ctx, user.ID, "USER_UPDATED", "user", user.ID.String(),
			fmt.Sprintf("User updated from role %s to %s", existingUser.Role, user.Role))
	}

	return nil
}

// DeleteUser deletes a user
func (us *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Check if user exists
	user, err := us.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Soft delete (set deleted_at)
	if err := us.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Log audit trail
	if us.auditService != nil {
		us.auditService.LogUserActivity(ctx, user.ID, "USER_DELETED", "user", user.ID.String(),
			fmt.Sprintf("User deleted with role: %s", user.Role))
	}

	return nil
}

// GetAllUsers retrieves all users with pagination
func (us *UserService) GetAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	users, err := us.userRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	return users, nil
}

// UpdateUserRole updates user role (role-based authorization)
func (us *UserService) UpdateUserRole(ctx context.Context, userID uuid.UUID, role models.UserRole) error {
	// Validate role
	switch role {
	case models.RoleCustomer, models.RoleAdmin, models.RoleTeller:
		// Valid role
	default:
		return fmt.Errorf("invalid role: %s", role)
	}

	// Check if user exists
	user, err := us.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	oldRole := user.Role

	// Update role using repository
	if err := us.userRepo.UpdateRole(ctx, userID, role); err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Log audit trail
	if us.auditService != nil {
		us.auditService.LogUserActivity(ctx, userID, "ROLE_UPDATED", "user", userID.String(),
			fmt.Sprintf("User role updated from %s to %s", oldRole, role))
	}

	return nil
}

// ValidateUserData validates user data
func (us *UserService) ValidateUserData(ctx context.Context, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	return nil
}

// CheckEmailAvailability checks if email is available
func (us *UserService) CheckEmailAvailability(ctx context.Context, email string) (bool, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	exists, err := us.userRepo.EmailExists(ctx, email)
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}
	return !exists, nil
}

// CheckUsernameAvailability checks if username is available
func (us *UserService) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	username = strings.TrimSpace(username)
	exists, err := us.userRepo.UsernameExists(ctx, username)
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}
	return !exists, nil
}

// Authentication methods

// AuthenticateUser authenticates a user with username/email and password
func (us *UserService) AuthenticateUser(ctx context.Context, usernameOrEmail, password string) (*models.User, error) {
	// Try to find user by username or email
	query := strings.TrimSpace(usernameOrEmail)

	// Try username first
	user, err := us.userRepo.GetByUsername(ctx, query)
	if err != nil {
		// Try email
		user, err = us.userRepo.GetByEmail(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("invalid credentials")
		}
	}

	// Verify password
	if !us.verifyPassword(password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Log successful authentication
	if us.auditService != nil {
		us.auditService.LogUserActivity(ctx, user.ID, "USER_LOGIN", "user", user.ID.String(), "User logged in successfully")
	}

	return user, nil
}

// ChangePassword changes user password
func (us *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	// Get user
	user, err := us.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify old password
	if !us.verifyPassword(oldPassword, user.PasswordHash) {
		return fmt.Errorf("invalid old password")
	}

	// Validate new password
	if err := models.ValidatePassword(newPassword); err != nil {
		return fmt.Errorf("new password validation failed: %w", err)
	}

	// Hash new password
	hashedPassword, err := us.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	// Update password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := us.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Log password change
	if us.auditService != nil {
		us.auditService.LogUserActivity(ctx, userID, "PASSWORD_CHANGED", "user", userID.String(), "Password changed successfully")
	}

	return nil
}

// Authorization methods

// HasRole checks if user has specific role
func (us *UserService) HasRole(user *models.User, role models.UserRole) bool {
	return user.Role == role
}

// HasAnyRole checks if user has any of the specified roles
func (us *UserService) HasAnyRole(user *models.User, roles ...models.UserRole) bool {
	for _, role := range roles {
		if user.Role == role {
			return true
		}
	}
	return false
}

// IsAdmin checks if user is admin
func (us *UserService) IsAdmin(user *models.User) bool {
	return user.Role == models.RoleAdmin
}

// IsTeller checks if user is teller
func (us *UserService) IsTeller(user *models.User) bool {
	return user.Role == models.RoleTeller
}

// IsCustomer checks if user is customer
func (us *UserService) IsCustomer(user *models.User) bool {
	return user.Role == models.RoleCustomer
}

// CanAccessAdminFeatures checks if user can access admin features
func (us *UserService) CanAccessAdminFeatures(user *models.User) bool {
	return user.CanAccessAdminFeatures()
}

// CanPerformTransaction checks if user can perform transactions
func (us *UserService) CanPerformTransaction(user *models.User) bool {
	return user.CanPerformTransaction()
}

// Helper methods

// hashPassword hashes a password using bcrypt
func (us *UserService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// verifyPassword verifies a password against its hash
func (us *UserService) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// validateCreateRequest validates user creation request
func (us *UserService) validateCreateRequest(req *models.UserCreateRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Username == "" {
		return fmt.Errorf("username is required")
	}

	if req.Email == "" {
		return fmt.Errorf("email is required")
	}

	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	// Validate password strength
	if err := models.ValidatePassword(req.Password); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate username format
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validUsername.MatchString(req.Username) {
		return fmt.Errorf("username can only contain letters, numbers, and underscores")
	}

	return nil
}

// GetUserCount returns total number of users
func (us *UserService) GetUserCount(ctx context.Context) (int64, error) {
	return us.userRepo.Count(ctx)
}
