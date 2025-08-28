package v1

import (
	"net/http"

	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/barannkoca/banking-backend/internal/services"
	"github.com/barannkoca/banking-backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserHandler handles user management requests
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUsers handles GET /api/v1/users - Get all users
func (h *UserHandler) GetUsers(c *gin.Context) {
	// Get users from service
	users, err := h.userService.GetAllUsers(c.Request.Context(), 0, 0) // Get all users
	if err != nil {
		logger.GetLogger().Error("Failed to get users",
			zap.Error(err),
			zap.String("ip", c.ClientIP()),
			zap.String("type", "user_list_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve users",
			"message": "Kullanıcı listesi alınamadı",
		})
		return
	}

	// Convert to response format
	var userResponses []*models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	logger.GetLogger().Info("Users retrieved successfully",
		zap.Int("count", len(userResponses)),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "user_list_success"),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Kullanıcılar başarıyla getirildi",
		"data":    userResponses,
	})
}

// GetUser handles GET /api/v1/users/{id} - Get user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	// Get user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.GetLogger().Warn("Invalid user ID format",
			zap.String("user_id", userIDStr),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "user_id_validation_error"),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "Geçersiz kullanıcı ID'si",
		})
		return
	}

	// Get current user from context (for authorization)
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Kimlik doğrulama gerekli",
		})
		return
	}

	// Check if user is requesting their own data or is admin
	currentUserIDStr := currentUserID.(string)
	currentUserUUID, _ := uuid.Parse(currentUserIDStr)

	// Get current user's role
	currentUser, err := h.userService.GetUserByID(c.Request.Context(), currentUserUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get current user",
			"message": "Mevcut kullanıcı bilgisi alınamadı",
		})
		return
	}

	// Authorization check: user can only access their own data unless they're admin
	if currentUserUUID != userID && currentUser.Role != models.RoleAdmin {
		logger.GetLogger().Warn("Unauthorized user access attempt",
			zap.String("current_user_id", currentUserIDStr),
			zap.String("requested_user_id", userID.String()),
			zap.String("ip", c.ClientIP()),
			zap.String("type", "user_access_unauthorized"),
		)

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "Bu kullanıcının bilgilerine erişim izniniz yok",
		})
		return
	}

	// Get user from service
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		logger.GetLogger().Error("Failed to get user",
			zap.String("user_id", userID.String()),
			zap.Error(err),
			zap.String("ip", c.ClientIP()),
			zap.String("type", "user_get_error"),
		)

		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"message": "Kullanıcı bulunamadı",
		})
		return
	}

	logger.GetLogger().Info("User retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "user_get_success"),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Kullanıcı başarıyla getirildi",
		"data":    user.ToResponse(),
	})
}

// UpdateUser handles PUT /api/v1/users/{id} - Update user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Get user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.GetLogger().Warn("Invalid user ID format",
			zap.String("user_id", userIDStr),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "user_id_validation_error"),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "Geçersiz kullanıcı ID'si",
		})
		return
	}

	// Get current user from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Kimlik doğrulama gerekli",
		})
		return
	}

	// Check authorization
	currentUserIDStr := currentUserID.(string)
	currentUserUUID, _ := uuid.Parse(currentUserIDStr)

	currentUser, err := h.userService.GetUserByID(c.Request.Context(), currentUserUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get current user",
			"message": "Mevcut kullanıcı bilgisi alınamadı",
		})
		return
	}

	// Authorization check: user can only update their own data unless they're admin
	if currentUserUUID != userID && currentUser.Role != models.RoleAdmin {
		logger.GetLogger().Warn("Unauthorized user update attempt",
			zap.String("current_user_id", currentUserIDStr),
			zap.String("requested_user_id", userID.String()),
			zap.String("ip", c.ClientIP()),
			zap.String("type", "user_update_unauthorized"),
		)

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "Bu kullanıcının bilgilerini güncelleme izniniz yok",
		})
		return
	}

	// Bind request body
	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Warn("Invalid update request",
			zap.String("user_id", userID.String()),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "user_update_validation_error"),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": "Geçersiz güncelleme verisi",
		})
		return
	}

	// Get existing user
	existingUser, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"message": "Kullanıcı bulunamadı",
		})
		return
	}

	// Update user fields
	if req.Username != "" {
		existingUser.Username = req.Username
	}
	if req.Email != "" {
		existingUser.Email = req.Email
	}
	if req.Role != "" {
		// Only admin can change roles
		if currentUser.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Role change not allowed",
				"message": "Rol değişikliği izni yok",
			})
			return
		}
		existingUser.Role = models.UserRole(req.Role)
	}

	// Update user
	err = h.userService.UpdateUser(c.Request.Context(), existingUser)
	if err != nil {
		logger.GetLogger().Error("Failed to update user",
			zap.String("user_id", userID.String()),
			zap.Error(err),
			zap.String("ip", c.ClientIP()),
			zap.String("type", "user_update_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user",
			"message": "Kullanıcı güncellenemedi",
		})
		return
	}

	logger.GetLogger().Info("User updated successfully",
		zap.String("user_id", userID.String()),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "user_update_success"),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Kullanıcı başarıyla güncellendi",
		"data":    existingUser.ToResponse(),
	})
}

// DeleteUser handles DELETE /api/v1/users/{id} - Delete user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Get user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.GetLogger().Warn("Invalid user ID format",
			zap.String("user_id", userIDStr),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "user_id_validation_error"),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "Geçersiz kullanıcı ID'si",
		})
		return
	}

	// Get current user from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Kimlik doğrulama gerekli",
		})
		return
	}

	// Check authorization
	currentUserIDStr := currentUserID.(string)
	currentUserUUID, _ := uuid.Parse(currentUserIDStr)

	currentUser, err := h.userService.GetUserByID(c.Request.Context(), currentUserUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get current user",
			"message": "Mevcut kullanıcı bilgisi alınamadı",
		})
		return
	}

	// Authorization check: user can only delete their own account unless they're admin
	if currentUserUUID != userID && currentUser.Role != models.RoleAdmin {
		logger.GetLogger().Warn("Unauthorized user deletion attempt",
			zap.String("current_user_id", currentUserIDStr),
			zap.String("requested_user_id", userID.String()),
			zap.String("ip", c.ClientIP()),
			zap.String("type", "user_delete_unauthorized"),
		)

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "Bu kullanıcıyı silme izniniz yok",
		})
		return
	}

	// Prevent admin from deleting themselves
	if currentUserUUID == userID && currentUser.Role == models.RoleAdmin {
		logger.GetLogger().Warn("Admin attempted to delete themselves",
			zap.String("user_id", userID.String()),
			zap.String("ip", c.ClientIP()),
			zap.String("type", "admin_self_delete_attempt"),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Cannot delete admin account",
			"message": "Admin hesabı silinemez",
		})
		return
	}

	// Delete user
	err = h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		logger.GetLogger().Error("Failed to delete user",
			zap.String("user_id", userID.String()),
			zap.Error(err),
			zap.String("ip", c.ClientIP()),
			zap.String("type", "user_delete_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete user",
			"message": "Kullanıcı silinemedi",
		})
		return
	}

	logger.GetLogger().Info("User deleted successfully",
		zap.String("user_id", userID.String()),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "user_delete_success"),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Kullanıcı başarıyla silindi",
	})
}
