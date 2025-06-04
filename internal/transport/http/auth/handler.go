package auth

import (
	"fmt"

	"errors" // Added for errors.Is

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth"
	serviceAuth "github.com/yi-tech/go-user-service/internal/service/auth" // Import for sentinel errors
	// userService "github.com/yi-tech/go-user-service/internal/service/user" // For userService.ErrUserNotFound if needed directly
	"github.com/yi-tech/go-user-service/internal/transport/http/response"
)

// Handler handles HTTP requests for authentication operations
type Handler struct {
	authService domainAuth.AuthService
	logger      *zap.Logger
}

// NewHandler creates a new authentication handler
func NewHandler(authService domainAuth.AuthService, logger *zap.Logger) *Handler {
	return &Handler{
		authService: authService,
		logger:      logger,
	}
}

// Login handles user login
// @Summary User login
// @Description Authenticate a user and return access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=LoginResponse} "Successfully authenticated"
// @Failure 400 {object} response.Response "Invalid request data"
// @Failure 401 {object} response.Response "Invalid email or password"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest // Use local DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request",
			zap.String("operation", "Login"),
			zap.Error(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Authenticate user
	tokenPair, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, serviceAuth.ErrInvalidCredentials) {
			h.logger.Info("Login attempt failed: invalid credentials", // This log is fine
				zap.String("operation", "Login"),
				zap.String("email", req.Email))
			response.Unauthorized(c, serviceAuth.ErrInvalidCredentials.Error())
			return // This return was correctly placed. The issue might be in test expectation or mock.
		}
		// For other (unexpected) errors, Error level is appropriate.
		h.logger.Error("Login error (unexpected)", // Clarified log message
			zap.String("operation", "Login"),
			zap.Error(err), // This err is not ErrInvalidCredentials here
			zap.String("email", req.Email))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	// Create response data
	loginData := LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    3600, // Placeholder for access token lifetime (e.g., 1 hour)
	}

	response.Success(c, loginData)
}

// RefreshToken handles refreshing an access token
// @Summary Refresh access token
// @Description Refresh an access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.Response{data=LoginResponse} "Token refreshed successfully"
// @Failure 400 {object} response.Response "Invalid request data"
// @Failure 401 {object} response.Response "Invalid or expired refresh token"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest // Use local DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh token request",
			zap.String("operation", "RefreshToken"),
			zap.Error(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Refresh token
	tokenPair, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, serviceAuth.ErrInvalidOrExpiredToken) {
			h.logger.Info("Refresh token failed: invalid or expired", // This log is fine
				zap.String("operation", "RefreshToken"),
				zap.Error(err)) // err here is ErrInvalidOrExpiredToken
			response.Unauthorized(c, serviceAuth.ErrInvalidOrExpiredToken.Error())
			return // This return was correctly placed.
		}
		// For other (unexpected) errors
		h.logger.Error("Failed to refresh token (unexpected)", // Clarified log message
			zap.String("operation", "RefreshToken"),
			zap.Error(err)) // This err is not ErrInvalidOrExpiredToken here
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	// Create response data
	responseData := LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    3600, // Placeholder for access token lifetime
	}

	response.Success(c, responseData)
}

// Logout handles user logout
// @Summary User logout
// @Description Invalidate the user's refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "Logged out successfully"
// @Failure 401 {object} response.Response "Authentication required"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	// Get user ID from context (assuming it's set by auth middleware)
	userIDRaw, exists := c.Get("userID") // Changed "user_id" to "userID"
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Assert userID to uuid.UUID
	userIDUUID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		h.logger.Error("Failed to assert user ID to uuid.UUID for logout",
			zap.String("operation", "Logout"),
			zap.Any("user_id_type", fmt.Sprintf("%T", userIDRaw)),
			zap.Any("user_id_value", userIDRaw))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	// Logout user
	err := h.authService.Logout(c.Request.Context(), userIDUUID)
	if err != nil {
		h.logger.Error("Failed to logout user",
			zap.String("operation", "Logout"),
			zap.Error(err),
			zap.String("user_id", userIDUUID.String()))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	response.Success(c, gin.H{"message": "Logged out successfully"})
}
