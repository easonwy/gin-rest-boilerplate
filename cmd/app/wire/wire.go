//go:build wireinject
// +build wireinject

package wire

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	authHandler "github.com/yi-tech/go-user-service/internal/auth/handler"
	authRepository "github.com/yi-tech/go-user-service/internal/auth/repository"
	authService "github.com/yi-tech/go-user-service/internal/auth/service"
	"github.com/yi-tech/go-user-service/internal/config"
	grpcServer "github.com/yi-tech/go-user-service/internal/grpc"
	"github.com/yi-tech/go-user-service/internal/provider"
	"github.com/yi-tech/go-user-service/internal/user"
	userHandler "github.com/yi-tech/go-user-service/internal/user/handler"
	"github.com/yi-tech/go-user-service/pkg/middleware"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProvideGRPCConfig provides the gRPC server configuration
func ProvideGRPCConfig(cfg *config.Config) *grpcServer.Config {
	return &grpcServer.Config{
		GRPCPort: cfg.GRPC.Port,
		HTTPPort: cfg.GRPC.Port + 1, // Use next port for HTTP gateway
	}
}

// App represents the main application structure.
type App struct {
	Router     *gin.Engine
	DB         *gorm.DB
	Config     *config.Config
	GRPCServer *grpcServer.Server // gRPC server instance
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
		authRepository.NewAuthRepository,
		middleware.AuthMiddleware,
		NewRouter,
		// gRPC setup
		ProvideGRPCConfig,
		grpcServer.NewServer,
		// Create the App with all dependencies
		wire.Struct(new(App), "*"),
	)
	return &App{}, nil // Wire will provide the actual implementation
}

// NewRouter creates a new Gin router and sets up routes.
func NewRouter(userHdl userHandler.UserHandler, authHdl authHandler.AuthHandler, authMiddleware gin.HandlerFunc, logger *zap.Logger) *gin.Engine {
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("/")
		{
			// User routes
			userGroup := public.Group("/users")
			{
				userGroup.POST("/register", userHdl.Register)
				userGroup.GET("", userHdl.GetUserByEmail) // GET /api/v1/users?email=test@example.com
				userGroup.GET("/:id", userHdl.GetUserByID)
			}

			// Auth routes
			authGroup := public.Group("/auth")
			{
				authGroup.POST("/login", authHdl.Login)
				authGroup.POST("/refresh", authHdl.RefreshToken)
			}
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(authMiddleware)
		{
			// User routes
			userGroup := protected.Group("/users")
			{
				userGroup.PUT("/:id", userHdl.UpdateUser)
				userGroup.DELETE("/:id", userHdl.DeleteUser)
			}

			// Profile routes (using the same user handler methods but with auth)
			profileGroup := protected.Group("/profile")
			{
				profileGroup.GET("", userHdl.GetUserByID) // Gets the current user's profile
				profileGroup.PUT("", userHdl.UpdateUser)    // Updates the current user's profile
			}
		}
	}

	return r
}
