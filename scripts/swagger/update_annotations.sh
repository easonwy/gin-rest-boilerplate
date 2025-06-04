#!/bin/bash

set -e

# Update user handler.go with Swagger annotations
cat > /Users/easonwu/Dev/personal/go-user-service/internal/transport/http/user/handler.go << 'EOF'
package user

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
	serviceUser "github.com/yi-tech/go-user-service/internal/service/user"
	"github.com/yi-tech/go-user-service/internal/transport/http/response"
	"go.uber.org/zap"
)

// UpdateRequest represents the request body for updating a user (kept inline for now)
type UpdateRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// Handler handles HTTP requests for user operations
type Handler struct {
	userService serviceUser.UserService
	logger      *zap.Logger
}

// NewHandler creates a new user handler
func NewHandler(userService serviceUser.UserService, logger *zap.Logger) *Handler {
	return &Handler{
		userService: userService,
		logger:      logger,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param request body UserRegisterRequest true "User registration information"
// @Success 201 {object} response.Response{data=UserResponse} "User registered successfully"
// @Failure 400 {object} response.Response "Invalid request data"
// @Failure 409 {object} response.Response "Email already exists"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid register request", zap.Error(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Call domain service with direct parameters
	newUser, err := h.userService.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		// Handle specific error cases
		if err.Error() == "user already exists" {
			response.Conflict(c, "Email already exists")
			return
		}
		
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to register user", zap.Error(err))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	// Use the response package with status code 201 (Created)
	c.JSON(http.StatusCreated, response.NewResponse(http.StatusCreated, "User registered successfully", toUserResponse(newUser)))
}

// GetUserByID handles retrieving a user by ID
// @Summary Get a user by ID
// @Description Retrieve a user's information by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=UserResponse} "User information"
// @Failure 400 {object} response.Response "Invalid user ID format"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	
	// Convert string ID to UUID
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID format")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userUUID)
	if err != nil {
		// Handle specific error cases
		if err.Error() == "user not found" {
			response.NotFound(c, "User not found")
			return
		}
		
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", idParam))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	response.Success(c, toUserResponse(user))
}

// GetUserByEmail handles retrieving a user by email
// @Summary Get a user by email
// @Description Retrieve a user's information by their email address
// @Tags users
// @Accept json
// @Produce json
// @Param email query string true "User email"
// @Success 200 {object} response.Response{data=UserResponse} "User information"
// @Failure 400 {object} response.Response "Email is required"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users [get]
func (h *Handler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.BadRequest(c, "Email is required")
		return
	}

	user, err := h.userService.GetByEmail(c.Request.Context(), email)
	if err != nil {
		// Handle specific error cases
		if err.Error() == "user not found" {
			response.NotFound(c, "User not found")
			return
		}
		
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	response.Success(c, toUserResponse(user))
}

// UpdateProfile handles updating a user's profile
// @Summary Update user profile
// @Description Update a user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UserUpdateRequest true "User update information"
// @Success 200 {object} response.Response{data=UserResponse} "User updated successfully"
// @Failure 400 {object} response.Response "Invalid request data or user ID format"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/{id} [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	idParam := c.Param("id")
	
	// Convert string ID to UUID
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID format")
		return
	}

	var req UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update request", zap.Error(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Get current user data
	currentUser, err := h.userService.GetByID(c.Request.Context(), userUUID)
	if err != nil {
		// Handle specific error cases
		if err.Error() == "user not found" {
			response.NotFound(c, "User not found")
			return
		}
		
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to get user for update", zap.Error(err), zap.String("user_id", idParam))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	// Apply updates (only if provided)
	updates := domainUser.UpdateUserParams{}
	
	if req.FirstName != nil {
		updates.FirstName = *req.FirstName
	}
	
	if req.LastName != nil {
		updates.LastName = *req.LastName
	}
	
	if req.Email != nil {
		updates.Email = *req.Email
	}

	// Update user
	updatedUser, err := h.userService.Update(c.Request.Context(), userUUID, updates)
	if err != nil {
		// Handle specific error cases
		if err.Error() == "email already in use" {
			response.Conflict(c, "Email already in use by another account")
			return
		}
		
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", idParam))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	response.Success(c, toUserResponse(updatedUser))
}

// UpdatePassword handles updating a user's password
// @Summary Update user password
// @Description Update a user's password
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdatePasswordRequest true "Password update information"
// @Success 200 {object} response.Response "Password updated successfully"
// @Failure 400 {object} response.Response "Invalid request data or user ID format"
// @Failure 401 {object} response.Response "Current password is incorrect"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/{id}/password [put]
func (h *Handler) UpdatePassword(c *gin.Context) {
	idParam := c.Param("id")
	
	// Convert string ID to UUID
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID format")
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid password update request", zap.Error(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Update password
	err = h.userService.UpdatePassword(c.Request.Context(), userUUID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		// Handle specific error cases
		switch err.Error() {
		case "user not found":
			response.NotFound(c, "User not found")
			return
		case "current password is incorrect":
			response.Unauthorized(c, "Current password is incorrect")
			return
		default:
			// Log the actual error for debugging but return a generic message
			h.logger.Error("Failed to update password", zap.Error(err), zap.String("user_id", idParam))
			response.InternalServerError(c, "Something went wrong. Please try again later.")
			return
		}
	}

	response.Success(c, gin.H{"message": "Password updated successfully"})
}

// DeleteUser handles deleting a user
// @Summary Delete a user
// @Description Delete a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response "User deleted successfully"
// @Failure 400 {object} response.Response "Invalid user ID format"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	
	// Convert string ID to UUID
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID format")
		return
	}

	// Delete user
	err = h.userService.Delete(c.Request.Context(), userUUID)
	if err != nil {
		// Handle specific error cases
		if err.Error() == "user not found" {
			response.NotFound(c, "User not found")
			return
		}
		
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", idParam))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	response.Success(c, gin.H{"message": "User deleted successfully"})
}

// Helper function to convert domain user to response DTO
func toUserResponse(user *domainUser.User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
EOF

# Update auth handler.go with Swagger annotations
cat > /Users/easonwu/Dev/personal/go-user-service/internal/transport/http/auth/handler.go << 'EOF'
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
EOF

# Make the script executable
chmod +x "$0"

echo "Swagger annotations added to handler files."