package v1

import (
	"net/http"

	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/barannkoca/banking-backend/internal/services"
	"github.com/barannkoca/banking-backend/pkg/logger"
	"github.com/barannkoca/banking-backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userService *services.UserService
	jwtConfig   *utils.JWTConfig
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtConfig:   utils.DefaultJWTConfig(),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.AuthRegisterRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Warn("Invalid registration request",
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "auth_validation_error"),
		)

		c.JSON(http.StatusBadRequest, models.NewAuthError(
			"Validation failed",
			"Geçersiz kayıt verileri",
			"VALIDATION_ERROR",
		))
		return
	}

	// Create user creation request
	userReq := &models.UserCreateRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}

	// Create user
	user, err := h.userService.CreateUser(c.Request.Context(), userReq)
	if err != nil {
		logger.GetLogger().Error("User creation failed",
			zap.String("username", req.Username),
			zap.String("email", req.Email),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "auth_creation_error"),
		)

		// Handle specific errors
		switch err.Error() {
		case "email already exists":
			c.JSON(http.StatusConflict, models.NewAuthError(
				"Email already exists",
				"Bu e-posta adresi zaten kullanımda",
				"EMAIL_EXISTS",
			))
		case "username already exists":
			c.JSON(http.StatusConflict, models.NewAuthError(
				"Username already exists",
				"Bu kullanıcı adı zaten kullanımda",
				"USERNAME_EXISTS",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.NewAuthError(
				"Registration failed",
				"Kayıt işlemi başarısız",
				"REGISTRATION_ERROR",
			))
		}
		return
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user, h.jwtConfig)
	if err != nil {
		logger.GetLogger().Error("Access token generation failed",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
			zap.String("type", "token_generation_error"),
		)
		c.JSON(http.StatusInternalServerError, models.NewAuthError(
			"Token generation failed",
			"Token oluşturma hatası",
			"TOKEN_ERROR",
		))
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user, h.jwtConfig)
	if err != nil {
		logger.GetLogger().Error("Refresh token generation failed",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
			zap.String("type", "token_generation_error"),
		)
		c.JSON(http.StatusInternalServerError, models.NewAuthError(
			"Token generation failed",
			"Token oluşturma hatası",
			"TOKEN_ERROR",
		))
		return
	}

	// Create response
	response := models.NewAuthResponse(
		accessToken,
		refreshToken,
		int64(h.jwtConfig.AccessTokenExpiry.Seconds()),
		user,
	)

	logger.GetLogger().Info("User registered successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "auth_success"),
	)

	c.JSON(http.StatusCreated, models.NewAuthSuccess(
		"Kullanıcı başarıyla kaydedildi",
		response,
	))
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.AuthLoginRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Warn("Invalid login request",
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "auth_validation_error"),
		)

		c.JSON(http.StatusBadRequest, models.NewAuthError(
			"Validation failed",
			"Geçersiz giriş verileri",
			"VALIDATION_ERROR",
		))
		return
	}

	// Authenticate user
	user, err := h.userService.AuthenticateUser(c.Request.Context(), req.UsernameOrEmail, req.Password)
	if err != nil {
		logger.GetLogger().Warn("Login failed",
			zap.String("username_or_email", req.UsernameOrEmail),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "auth_failed"),
		)

		c.JSON(http.StatusUnauthorized, models.NewAuthError(
			"Invalid credentials",
			"Geçersiz kullanıcı adı/e-posta veya şifre",
			"INVALID_CREDENTIALS",
		))
		return
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user, h.jwtConfig)
	if err != nil {
		logger.GetLogger().Error("Access token generation failed",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
			zap.String("type", "token_generation_error"),
		)
		c.JSON(http.StatusInternalServerError, models.NewAuthError(
			"Token generation failed",
			"Token oluşturma hatası",
			"TOKEN_ERROR",
		))
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user, h.jwtConfig)
	if err != nil {
		logger.GetLogger().Error("Refresh token generation failed",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
			zap.String("type", "token_generation_error"),
		)
		c.JSON(http.StatusInternalServerError, models.NewAuthError(
			"Token generation failed",
			"Token oluşturma hatası",
			"TOKEN_ERROR",
		))
		return
	}

	// Create response
	response := models.NewAuthResponse(
		accessToken,
		refreshToken,
		int64(h.jwtConfig.AccessTokenExpiry.Seconds()),
		user,
	)

	logger.GetLogger().Info("User logged in successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "auth_success"),
	)

	c.JSON(http.StatusOK, models.NewAuthSuccess(
		"Giriş başarılı",
		response,
	))
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.AuthRefreshRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Warn("Invalid refresh token request",
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "auth_validation_error"),
		)

		c.JSON(http.StatusBadRequest, models.NewAuthError(
			"Validation failed",
			"Geçersiz refresh token verisi",
			"VALIDATION_ERROR",
		))
		return
	}

	// Refresh token pair
	accessToken, refreshToken, err := utils.RefreshTokenPair(req.RefreshToken, h.jwtConfig)
	if err != nil {
		logger.GetLogger().Warn("Token refresh failed",
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "token_refresh_error"),
		)

		c.JSON(http.StatusUnauthorized, models.NewAuthError(
			"Invalid refresh token",
			"Geçersiz refresh token",
			"INVALID_REFRESH_TOKEN",
		))
		return
	}

	// Create response
	response := models.NewAuthRefreshResponse(
		accessToken,
		refreshToken,
		int64(h.jwtConfig.AccessTokenExpiry.Seconds()),
	)

	logger.GetLogger().Info("Token refreshed successfully",
		zap.String("ip", c.ClientIP()),
		zap.String("type", "token_refresh_success"),
	)

	c.JSON(http.StatusOK, models.NewAuthSuccess(
		"Token başarıyla yenilendi",
		response,
	))
}

// Logout handles user logout (optional - for token blacklisting)
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewAuthError(
			"Authentication required",
			"Kimlik doğrulama gerekli",
			"AUTH_REQUIRED",
		))
		return
	}

	logger.GetLogger().Info("User logged out",
		zap.String("user_id", userID.(string)),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "auth_logout"),
	)

	c.JSON(http.StatusOK, models.NewAuthSuccess(
		"Çıkış başarılı",
		gin.H{"message": "User logged out successfully"},
	))
}
