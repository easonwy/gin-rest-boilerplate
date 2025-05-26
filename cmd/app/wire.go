//go:build wireinject
// +build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	authHandler "github.com/tapas/go-user-service/internal/auth/handler"
	authRepository "github.com/tapas/go-user-service/internal/auth/repository"
	authService "github.com/tapas/go-user-service/internal/auth/service"
	"github.com/tapas/go-user-service/internal/config"
	"github.com/tapas/go-user-service/internal/provider"
	"github.com/tapas/go-user-service/internal/user"
	userHandler "github.com/tapas/go-user-service/internal/user/handler"
	"github.com/tapas/go-user-service/pkg/middleware"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App represents the main application structure.
type App struct {
	Router *gin.Engine
	DB     *gorm.DB
	Config *config.Config
}

// InitializeApp creates the application dependencies.
func InitializeApp() (*App, error) {
	wire.Build(
		provider.ProvideConfig,
		provider.ProvideDatabase,
		provider.ProvideRedisClient,
		provider.ProvideLogger,
		user.ProvideUserRepository,
		user.ProvideUserService,
		user.ProvideUserHandler,
		authService.NewAuthService,
		authHandler.NewAuthHandler,
		middleware.AuthMiddleware,
		authRepository.NewAuthRepository,
		NewRouter,
		wire.Struct(new(App), "Router", "DB", "Config"),
	)
	return &App{}, nil // Wire will provide the actual implementation
}

// NewRouter creates a new Gin router and sets up routes.
func NewRouter(userHdl userHandler.UserHandler, authHandler authHandler.AuthHandler, authMiddleware gin.HandlerFunc, logger *zap.Logger) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.LoggingMiddleware(logger))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// API v1 group with base path
	api := r.Group("/api/v1")

	// 公开的用户路由
	api.POST("/users/register", userHdl.Register)

	// 需要认证的用户路由
	userGroup := api.Group("/users")
	userGroup.Use(authMiddleware) // Apply JWT auth middleware
	{
		userGroup.GET("/:id", userHdl.GetUserByID)
		userGroup.GET("/", userHdl.GetUserByEmail)
		userGroup.PUT("/:id", userHdl.UpdateUser)
		userGroup.DELETE("/:id", userHdl.DeleteUser)
		// TODO: Add other user routes
	}

	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh-token", authHandler.RefreshToken)
		// TODO: Add other auth routes
	}

	authProtectedGroup := api.Group("/auth")
	authProtectedGroup.Use(authMiddleware) // Apply JWT auth middleware
	{
		authProtectedGroup.POST("/logout", authHandler.Logout)
	}

	return r
}
