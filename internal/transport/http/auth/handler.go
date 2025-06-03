package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth"
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
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest // Use local DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Authenticate user
	tokenPair, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Info("Login attempt failed", zap.Error(err), zap.String("email", req.Email))
		
		// Handle specific error cases with user-friendly messages
		if err.Error() == "invalid credentials" {
			response.Unauthorized(c, "Invalid email or password")
			return
		}
		
		// Log the actual error for debugging but return a generic message to the user
		h.logger.Error("Login error", zap.Error(err), zap.String("email", req.Email))
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
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest // Use local DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh token request", zap.Error(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Refresh token
	tokenPair, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		// Handle specific error cases
		if err.Error() == "invalid or expired refresh token" {
			response.Unauthorized(c, "Invalid or expired refresh token")
			return
		}
		
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to refresh token", zap.Error(err))
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
func (h *Handler) Logout(c *gin.Context) {
	// Get user ID from context (assuming it's set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Assert userID to uuid.UUID
	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.logger.Error("Failed to assert user ID to uuid.UUID", zap.Any("user_id_type", fmt.Sprintf("%T", userID)), zap.Any("user_id_value", userID))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	// Logout user
	err := h.authService.Logout(c.Request.Context(), userIDUUID)
	if err != nil {
		h.logger.Error("Failed to logout user", zap.Error(err), zap.String("user_id", userIDUUID.String()))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	response.Success(c, gin.H{"message": "Logged out successfully"})
}
