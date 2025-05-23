package handler

import (
	"net/http"
	"strconv"

	"github.com/example/go-user-service/internal/user/dto"
	"github.com/example/go-user-service/internal/user/service"
	"github.com/example/go-user-service/pkg/response"
	"github.com/gin-gonic/gin"
)

// UserHandler defines the interface for user handlers.
type UserHandler interface {
	Register(c *gin.Context)
	GetUserByID(c *gin.Context)
	GetUserByEmail(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
	// TODO: Add other user related handler methods
}

type userHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(userService service.UserService) UserHandler {
	return &userHandler{userService: userService}
}

// Register handles user registration requests.
func (h *userHandler) Register(c *gin.Context) {
	var req dto.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body") // Corrected Error call
		return
	}

	user, err := h.userService.RegisterUser(&req) // Corrected Service method call
	if err != nil {
		// TODO: Handle specific service errors (e.g., username/email already exists)
		response.Error(c, http.StatusInternalServerError, "Failed to register user") // Corrected Error call
		return
	}

	response.Success(c, user) // Corrected Success call
}

// GetUserByID handles requests to get a user by ID.
func (h *userHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		// TODO: Handle specific service errors (e.g., user not found)
		response.Error(c, http.StatusInternalServerError, "Failed to get user")
		return
	}

	if user == nil {
		response.NotFound(c, "User not found")
		return
	}

	// Convert model.User to dto.UserResponse
	userResp := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	response.Success(c, userResp)
}

// GetUserByEmail handles requests to get a user by email.
func (h *userHandler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.BadRequest(c, "Email query parameter is required")
		return
	}

	user, err := h.userService.GetUserByEmail(email)
	if err != nil {
		// TODO: Handle specific service errors (e.g., user not found)
		response.Error(c, http.StatusInternalServerError, "Failed to get user")
		return
	}

	if user == nil {
		response.NotFound(c, "User not found")
		return
	}

	// Convert model.User to dto.UserResponse
	userResp := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	response.Success(c, userResp)
}

// UpdateUser handles requests to update a user.
func (h *userHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	user, err := h.userService.UpdateUser(uint(id), &req)
	if err != nil {
		// TODO: Handle specific service errors (e.g., user not found, validation errors)
		response.Error(c, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// Convert model.User to dto.UserResponse
	userResp := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	response.Success(c, userResp)
}

// DeleteUser handles requests to delete a user.
func (h *userHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	err = h.userService.DeleteUser(uint(id))
	if err != nil {
		// TODO: Handle specific service errors (e.g., user not found)
		response.Error(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	response.Success(c, gin.H{"message": "User deleted successfully"})
}
