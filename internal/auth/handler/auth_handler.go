package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yi-tech/go-user-service/internal/auth/model"
	"github.com/yi-tech/go-user-service/internal/auth/service"
	"github.com/yi-tech/go-user-service/pkg/response"
) // Import unified response package

// AuthHandler 定义认证服务 Handler 接口
type AuthHandler interface {
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
}

// authHandler 认证服务 Handler 实现结构体
type authHandler struct {
	authService service.AuthService
}

// NewAuthHandler 创建新的认证服务 Handler 实例
func NewAuthHandler(authService service.AuthService) AuthHandler {
	return &authHandler{
		authService: authService,
	}
}

// Login 处理用户登录 HTTP 请求
// @Summary User login
// @Description Authenticate a user and return access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=model.LoginResponse} "Successfully authenticated"
// @Failure 400 {object} response.Response "Invalid request body"
// @Failure 401 {object} response.Response "Invalid credentials"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /auth/login [post]
func (h *authHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		// TODO: Handle specific errors from service and map to appropriate HTTP status codes
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// RefreshToken 处理刷新令牌 HTTP 请求
// @Summary Refresh access token
// @Description Refresh an expired access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} response.Response{data=model.LoginResponse} "Successfully refreshed token"
// @Failure 400 {object} response.Response "Invalid request body"
// @Failure 401 {object} response.Response "Invalid or expired refresh token"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /auth/refresh-token [post]
func (h *authHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		// TODO: Handle specific errors from service and map to appropriate HTTP status codes
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// Logout handles user logout HTTP requests
// @Summary User logout
// @Description Invalidate the user's refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=object{message=string}} "Successfully logged out"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *authHandler) Logout(c *gin.Context) {
	// Get user ID from context (set by JWT middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User ID not found in context")
		return
	}

	// Convert userID to uint
	id, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID type in context")
		return
	}

	err := h.authService.Logout(c.Request.Context(), id)
	if err != nil {
		// TODO: Handle specific errors from service
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Logged out successfully"})
}
