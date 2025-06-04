package user

import (
	"errors" // Added for errors.Is
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
	realServiceUser "github.com/yi-tech/go-user-service/internal/service/user" // Renamed to avoid conflict with package name 'user'
	"github.com/yi-tech/go-user-service/internal/transport/http/response"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for user operations
type Handler struct {
	userService realServiceUser.UserService // Use the new alias
	logger      *zap.Logger
}

// NewHandler creates a new user handler
func NewHandler(userService realServiceUser.UserService, logger *zap.Logger) *Handler {
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
		h.logger.Warn("Invalid register request",
			zap.String("operation", "Register"),
			zap.Error(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Call domain service with direct parameters
	newUser, err := h.userService.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if errors.Is(err, realServiceUser.ErrUserAlreadyExists) {
			// No need to log ErrUserAlreadyExists as error, it's a known business logic case
			response.Conflict(c, realServiceUser.ErrUserAlreadyExists.Error())
			return
		}
		h.logger.Error("Failed to register user",
			zap.String("operation", "Register"),
			zap.Error(err),
			zap.String("email", req.Email)) // Add email for context
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
		if errors.Is(err, realServiceUser.ErrUserNotFound) {
			response.NotFound(c, realServiceUser.ErrUserNotFound.Error())
			return
		}
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to get user by ID",
			zap.String("operation", "GetUserByID"),
			zap.Error(err),
			zap.String("user_id", idParam))
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
		if errors.Is(err, realServiceUser.ErrUserNotFound) {
			response.NotFound(c, realServiceUser.ErrUserNotFound.Error())
			return
		}
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to get user by email",
			zap.String("operation", "GetUserByEmail"),
			zap.Error(err),
			zap.String("email", email))
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
		h.logger.Warn("Invalid update profile request",
			zap.String("operation", "UpdateProfile"),
			zap.Error(err),
			zap.String("user_id", idParam))
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Get current user data
	_, err = h.userService.GetByID(c.Request.Context(), userUUID) // Check if user exists before update
	if err != nil {
		if errors.Is(err, realServiceUser.ErrUserNotFound) {
			response.NotFound(c, realServiceUser.ErrUserNotFound.Error())
			return
		}
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to get user for update",
			zap.String("operation", "UpdateProfile"),
			zap.Error(err),
			zap.String("user_id", idParam))
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
		if errors.Is(err, realServiceUser.ErrUserNotFound) { // Should ideally not happen if GetByID above passed
			response.NotFound(c, realServiceUser.ErrUserNotFound.Error())
			return
		}
		if errors.Is(err, realServiceUser.ErrEmailInUse) {
			response.Conflict(c, realServiceUser.ErrEmailInUse.Error())
			return
		}
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to update user",
			zap.String("operation", "UpdateProfile"),
			zap.Error(err),
			zap.String("user_id", idParam))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	// Return updated user data
	response.Success(c, UserResponse{
		ID:        updatedUser.ID.String(),
		Email:     updatedUser.Email,
		FirstName: updatedUser.FirstName,
		LastName:  updatedUser.LastName,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	})
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
// @Router /users/{id}/password [patch]
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
		h.logger.Warn("Invalid password update request",
			zap.String("operation", "UpdatePassword"),
			zap.Error(err),
			zap.String("user_id", idParam))
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Update password
	err = h.userService.UpdatePassword(c.Request.Context(), userUUID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, realServiceUser.ErrUserNotFound) {
			response.NotFound(c, realServiceUser.ErrUserNotFound.Error())
			return
		}
		if errors.Is(err, realServiceUser.ErrIncorrectPassword) {
			response.Unauthorized(c, realServiceUser.ErrIncorrectPassword.Error())
			return
		}
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to update password",
			zap.String("operation", "UpdatePassword"),
			zap.Error(err),
			zap.String("user_id", idParam))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}
	// Removed extra closing brace here

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
	err = h.userService.DeleteUser(c.Request.Context(), userUUID)
	if err != nil {
		if errors.Is(err, realServiceUser.ErrUserNotFound) {
			response.NotFound(c, realServiceUser.ErrUserNotFound.Error())
			return
		}
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to delete user",
			zap.String("operation", "DeleteUser"),
			zap.Error(err),
			zap.String("user_id", idParam))
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

// GetProfile handles retrieving the current user's profile
// @Summary Get current user profile
// @Description Retrieve the current user's profile information
// @Tags profile
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=UserResponse} "User profile information"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Convert to UUID
	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.logger.Error("Failed to convert userID to UUID",
			zap.String("operation", "GetProfile"),
			zap.Any("userID_type", userID), // Log the type of userID
			zap.Any("userID_value", userID)) // Log the value of userID
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	// Get user data
	user, err := h.userService.GetByID(c.Request.Context(), userUUID)
	if err != nil {
		if errors.Is(err, realServiceUser.ErrUserNotFound) {
			// No need to log ErrUserNotFound as error, it's a known case for GetProfile
			response.NotFound(c, realServiceUser.ErrUserNotFound.Error()) // User from token not found
			return
		}
		// Log the actual error for debugging but return a generic message
		h.logger.Error("Failed to get user profile",
			zap.String("operation", "GetProfile"),
			zap.Error(err),
			zap.String("user_id", userUUID.String()))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	response.Success(c, toUserResponse(user))
}

// UpdateCurrentUserProfile handles updating the currently authenticated user's profile
// @Summary Update current user profile
// @Description Update the currently authenticated user's profile information
// @Tags profile
// @Accept json
// @Produce json
// @Param request body UpdateCurrentUserProfileRequest true "User profile update information"
// @Success 200 {object} response.Response{data=UserResponse} "Profile updated successfully"
// @Failure 400 {object} response.Response "Invalid request data"
// @Failure 401 {object} response.Response "User not authenticated"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /profile [put]
func (h *Handler) UpdateCurrentUserProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDRaw, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Convert to UUID
	userUUID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		h.logger.Error("Failed to convert userID to UUID from context for profile update",
			zap.String("operation", "UpdateCurrentUserProfile"),
			zap.Any("userID_type", userIDRaw),
			zap.Any("userID_value", userIDRaw))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	var req UpdateCurrentUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update current user profile request",
			zap.String("operation", "UpdateCurrentUserProfile"),
			zap.Error(err),
			zap.String("user_id", userUUID.String()))
		response.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

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

	// Call the existing Update method in the service
	updatedUser, err := h.userService.Update(c.Request.Context(), userUUID, updates)
	if err != nil {
		if errors.Is(err, realServiceUser.ErrUserNotFound) { // Should not happen if userID from token is valid and user exists
			response.NotFound(c, realServiceUser.ErrUserNotFound.Error())
			return
		}
		if errors.Is(err, realServiceUser.ErrEmailInUse) {
			response.Conflict(c, realServiceUser.ErrEmailInUse.Error())
			return
		}
		h.logger.Error("Failed to update current user profile",
			zap.String("operation", "UpdateCurrentUserProfile"),
			zap.Error(err),
			zap.String("user_id", userUUID.String()))
		response.InternalServerError(c, "Something went wrong. Please try again later.")
		return
	}

	response.Success(c, toUserResponse(updatedUser))
}
