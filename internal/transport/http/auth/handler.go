package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Authenticate user
	// Call updated service method
	tokenPair, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "invalid credentials" { // This error string might change based on service impl
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		h.logger.Error("Failed to login user", zap.Error(err), zap.String("email", req.Email))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    3600, // Placeholder for access token lifetime (e.g., 1 hour)
	})
}

// RefreshToken handles refreshing an access token
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest // Use local DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Refresh token
	// Call updated service method
	tokenPair, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "invalid or expired refresh token" { // This error string might change
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}
		h.logger.Error("Failed to refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    3600, // Placeholder for access token lifetime
	})
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	// Get user ID from context (assuming it's set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Assert userID to uuid.UUID
	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.logger.Error("Failed to assert user ID to uuid.UUID", zap.Any("user_id_type", fmt.Sprintf("%T", userID)), zap.Any("user_id_value", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error, user ID type mismatch"})
		return
	}

	// Logout user
	err := h.authService.Logout(c.Request.Context(), userIDUUID)
	if err != nil {
		h.logger.Error("Failed to logout user", zap.Error(err), zap.String("user_id", userIDUUID.String()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
