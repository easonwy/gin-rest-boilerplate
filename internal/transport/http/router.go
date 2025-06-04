package http

import (
	"github.com/gin-gonic/gin"
	"github.com/yi-tech/go-user-service/internal/domain/auth"
	"github.com/yi-tech/go-user-service/internal/middleware"
	authHandler "github.com/yi-tech/go-user-service/internal/transport/http/auth"
	"github.com/yi-tech/go-user-service/internal/transport/http/response"
	userHandler "github.com/yi-tech/go-user-service/internal/transport/http/user"
	"go.uber.org/zap"
)

// SetupRouter configures the Gin router with all routes
func SetupRouter(
	router *gin.Engine,
	userHandler *userHandler.Handler,
	authHandler *authHandler.Handler,
	authService auth.AuthService,
	logger *zap.Logger,
) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("/")
		{
			// User routes
			userGroup := public.Group("/users")
			{
				userGroup.POST("/register", userHandler.Register)
				userGroup.GET("", userHandler.GetUserByEmail)
				userGroup.GET("/:id", userHandler.GetUserByID)
			}

			// Auth routes
			authGroup := public.Group("/auth")
			{
				authGroup.POST("/login", authHandler.Login)
				authGroup.POST("/refresh", authHandler.RefreshToken)
				authGroup.POST("/logout", authHandler.Logout)
			}
		}

		// Protected routes (require authentication)
		authMiddleware := middleware.AuthMiddleware(authService, logger)
		protected := v1.Group("/")
		protected.Use(authMiddleware)
		{
			// User routes
			userGroup := protected.Group("/users")
			{
				userGroup.PUT("/:id", userHandler.UpdateProfile) // This remains PUT for admin/specific user update
				userGroup.PATCH("/:id/password", userHandler.UpdatePassword)
				userGroup.DELETE("/:id", userHandler.DeleteUser)
			}

			// Profile routes
			profileGroup := protected.Group("/profile")
			{
				profileGroup.GET("", userHandler.GetProfile)
				profileGroup.PUT("", userHandler.UpdateCurrentUserProfile)
			}
		}
	}
}

// NewRouter creates a new Gin router and sets up routes
func NewRouter(
	userHandler *userHandler.Handler,
	authHandler *authHandler.Handler,
	authService auth.AuthService,
	logger *zap.Logger,
) *gin.Engine {
	router := gin.New()

	// Use middleware
	router.Use(gin.Recovery())

	// Setup routes
	SetupRouter(router, userHandler, authHandler, authService, logger)

	return router
}
