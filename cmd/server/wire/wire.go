//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/yi-tech/go-user-service/internal/config"
	"github.com/yi-tech/go-user-service/internal/provider"
	auth "github.com/yi-tech/go-user-service/internal/domain/auth"
	auth2 "github.com/yi-tech/go-user-service/internal/repository/auth"
	user2 "github.com/yi-tech/go-user-service/internal/repository/user"
	auth3 "github.com/yi-tech/go-user-service/internal/service/auth"
	user3 "github.com/yi-tech/go-user-service/internal/service/user"
	auth5 "github.com/yi-tech/go-user-service/internal/transport/grpc/auth"
	user5 "github.com/yi-tech/go-user-service/internal/transport/grpc/user"
	auth4 "github.com/yi-tech/go-user-service/internal/transport/http/auth"
	user4 "github.com/yi-tech/go-user-service/internal/transport/http/user"
	grpc "github.com/yi-tech/go-user-service/internal/transport/grpc"
	http "github.com/yi-tech/go-user-service/internal/transport/http"
	"github.com/yi-tech/go-user-service/internal/middleware"
)

// ProvideGRPCConfig provides the gRPC server configuration
func ProvideGRPCConfig(cfg *config.Config) *grpc.Config {
	return &grpc.Config{
		GRPCPort: cfg.GRPC.Port,
		HTTPPort: cfg.GRPC.Port + 1, // Use next port for HTTP gateway to avoid conflict with main HTTP server
	}
}

// ProvideGRPCServer creates a new gRPC server
func ProvideGRPCServer(userService user3.UserService, authService auth.AuthService, logger *zap.Logger, cfg *grpc.Config) *grpc.Server {
	return grpc.NewServer(userService, authService, logger, cfg)
}

// App represents the main application structure.
type App struct {
	HTTPServer *http.Server // HTTP server (Gin) instance
	GRPCServer *grpc.Server // gRPC server instance
	DB         *gorm.DB
	Config     *config.Config
	Logger     *zap.Logger
}

// InitializeApp creates the application dependencies.
func InitializeApp() (*App, error) {
	wire.Build(
		provider.ProvideConfig,
		provider.ProvideLogger, // Now takes config as parameter
		provider.ProvideDatabase,
		provider.ProvideRedisClient,
		ProvideUserRepository,
		ProvideAuthRepository,
		ProvideUserService,
		ProvideAuthService,
		ProvideUserHttpHandler,
		ProvideAuthHttpHandler,
		ProvideRouter,
		ProvideGRPCConfig,
		ProvideGRPCServer,
		ProvideHTTPServer,
		wire.Struct(new(App), "*"),
	)

	return &App{}, nil // Wire will provide the actual implementation
}

// Provider functions for repositories
func ProvideUserRepository(db *gorm.DB) user2.UserRepository {
	return user2.NewUserRepository(db)
}

func ProvideAuthRepository(redis *redis.Client) auth2.AuthRepository {
	return auth2.NewAuthRepository(redis)
}

// Provider functions for services
func ProvideUserService(repo user2.UserRepository) user3.UserService {
	return user3.NewUserService(repo)
}

func ProvideAuthService(userService user3.UserService, authRepo auth2.AuthRepository, cfg *config.Config) auth.AuthService {
	return auth3.NewService(userService, authRepo, cfg)
}

// Provider functions for HTTP handlers
func ProvideUserHttpHandler(userService user3.UserService, logger *zap.Logger) *user4.Handler {
	return user4.NewHandler(userService, logger)
}

func ProvideAuthHttpHandler(authService auth.AuthService, logger *zap.Logger) *auth4.Handler {
	return auth4.NewHandler(authService, logger)
}

// Provider functions for gRPC handlers
func ProvideUserGrpcHandler(userService user3.UserService, logger *zap.Logger) *user5.Handler {
	return user5.NewHandler(userService, logger)
}

func ProvideAuthGrpcHandler(authService auth.AuthService, logger *zap.Logger) *auth5.Handler {
	return auth5.NewHandler(authService, logger)
}

// Provider function for middleware
func ProvideAuthMiddleware(authService auth.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return middleware.AuthMiddleware(authService, logger)
}

// Provider function for router
func ProvideRouter(userHandler *user4.Handler, authHandler *auth4.Handler, authService auth.AuthService, logger *zap.Logger) *gin.Engine {
	return http.NewRouter(userHandler, authHandler, authService, logger)
}

// ProvideHTTPServer creates a new HTTP server
func ProvideHTTPServer(router *gin.Engine, cfg *config.Config) *http.Server {
	return http.NewServer(router, cfg)
}


