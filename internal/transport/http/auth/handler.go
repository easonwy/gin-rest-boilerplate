package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yi-tech/go-user-service/internal/domain/auth"
	"github.com/yi-tech/go-user-service/internal/domain/auth/dto"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for authentication operations
type Handler struct {
	authService auth.AuthService
	logger      *zap.Logger
}

// NewHandler creates a new authentication handler
func NewHandler(authService auth.AuthService, logger *zap.Logger) *Handler {
	return &Handler{
		authService: authService,
		logger:      logger,
	}
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Authenticate user
	tokenPair, err := h.authService.Login(req)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "invalid credentials" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		h.logger.Error("Failed to login user", zap.Error(err), zap.String("email", req.Email))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
	})
}

// RefreshToken handles refreshing an access token
func (h *Handler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Refresh token
	tokenPair, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "invalid or expired refresh token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}
		h.logger.Error("Failed to refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
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

	// Convert to uint
	userIDUint, ok := userID.(uint)
	if !ok {
		h.logger.Error("Failed to convert user ID to uint", zap.Any("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Logout user
	err := h.authService.Logout(c.Request.Context(), userIDUint)
	if err != nil {
		h.logger.Error("Failed to logout user", zap.Error(err), zap.Uint("user_id", userIDUint))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
