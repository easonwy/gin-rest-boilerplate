package user

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yi-tech/go-user-service/internal/domain/user"
	userService "github.com/yi-tech/go-user-service/internal/service/user"
	"go.uber.org/zap"
)

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// UpdateRequest represents the request body for updating a user
type UpdateRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// PasswordUpdateRequest represents the request body for updating a password
type PasswordUpdateRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// UserResponse represents the response body for user data
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Handler handles HTTP requests for user operations
type Handler struct {
	userService userService.UserService
	logger      *zap.Logger
}

// NewHandler creates a new user handler
func NewHandler(userService userService.UserService, logger *zap.Logger) *Handler {
	return &Handler{
		userService: userService,
		logger:      logger,
	}
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid register request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Call domain service with direct parameters
	newUser, err := h.userService.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		h.logger.Error("Failed to register user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, toUserResponse(newUser))
}

// GetUserByID handles retrieving a user by ID
func (h *Handler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	
	// Convert string ID to UUID
	// Note: The domain model currently uses uint for ID, but the service interface expects UUID
	// This is a temporary solution during the transition to UUID
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userUUID)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", idParam))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

// GetUserByEmail handles retrieving a user by email
func (h *Handler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter is required"})
		return
	}

	user, err := h.userService.GetByEmail(c.Request.Context(), email)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

// UpdateUser handles updating a user's details
func (h *Handler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	
	// Convert string ID to UUID
	// Note: The domain model currently uses uint for ID, but the service interface expects UUID
	// This is a temporary solution during the transition to UUID
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update user with domain service
	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), userUUID, req.FirstName, req.LastName)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", idParam))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(updatedUser))
}

// UpdatePassword handles updating a user's password
func (h *Handler) UpdatePassword(c *gin.Context) {
	idParam := c.Param("id")
	
	// Convert string ID to UUID
	// Note: The domain model currently uses uint for ID, but the service interface expects UUID
	// This is a temporary solution during the transition to UUID
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var req PasswordUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid password update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update password with domain service
	err = h.userService.UpdatePassword(c.Request.Context(), userUUID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error("Failed to update password", zap.Error(err), zap.String("user_id", idParam))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// DeleteUser handles deleting a user
func (h *Handler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	
	// Convert string ID to UUID
	// Note: The domain model currently uses uint for ID, but the service interface expects UUID
	// This is a temporary solution during the transition to UUID
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), userUUID)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", idParam))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetProfile handles retrieving the current user's profile
func (h *Handler) GetProfile(c *gin.Context) {
	// The user ID should be set by the auth middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert interface{} to UUID
	// Note: The auth middleware may still be setting a uint or string ID, but the service interface expects UUID
	// This is a temporary solution during the transition to UUID
	var userUUID uuid.UUID
	switch v := userID.(type) {
	case string:
		parsedUUID, err := uuid.Parse(v)
		if err != nil {
			h.logger.Error("Invalid user ID format in context", zap.Error(err), zap.String("user_id", v))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		userUUID = parsedUUID
	case uuid.UUID:
		userUUID = v
	default:
		h.logger.Error("Invalid user ID type in context", zap.Any("user_id_type", fmt.Sprintf("%T", userID)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userUUID)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error("Failed to get user profile", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

// UpdateProfile handles updating the current user's profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	// The user ID should be set by the auth middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert interface{} to UUID
	// Note: The auth middleware may still be setting a uint or string ID, but the service interface expects UUID
	// This is a temporary solution during the transition to UUID
	var userUUID uuid.UUID
	switch v := userID.(type) {
	case string:
		parsedUUID, err := uuid.Parse(v)
		if err != nil {
			h.logger.Error("Invalid user ID format in context", zap.Error(err), zap.String("user_id", v))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		userUUID = parsedUUID
	case uuid.UUID:
		userUUID = v
	default:
		h.logger.Error("Invalid user ID type in context", zap.Any("user_id_type", fmt.Sprintf("%T", userID)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update user with domain service
	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), userUUID, req.FirstName, req.LastName)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error("Failed to update user profile", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(updatedUser))
}

// toUserResponse converts a domain user model to a response object
func toUserResponse(user *user.User) UserResponse {
	return UserResponse{
		ID:        fmt.Sprintf("%d", user.ID),
		Email:     user.Email,
		FirstName: user.Username, // Using Username as FirstName since our domain model doesn't have FirstName/LastName
		LastName:  "", // No LastName in the domain model
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
