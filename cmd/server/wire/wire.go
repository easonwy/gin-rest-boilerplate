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
	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth"
	"github.com/yi-tech/go-user-service/internal/middleware"
	"github.com/yi-tech/go-user-service/internal/provider"
	repoAuth "github.com/yi-tech/go-user-service/internal/repository/auth"
	repoUser "github.com/yi-tech/go-user-service/internal/repository/user"
	serviceAuth "github.com/yi-tech/go-user-service/internal/service/auth"
	serviceUser "github.com/yi-tech/go-user-service/internal/service/user"
	grpc "github.com/yi-tech/go-user-service/internal/transport/grpc"
	grpcAuth "github.com/yi-tech/go-user-service/internal/transport/grpc/auth"
	grpcUser "github.com/yi-tech/go-user-service/internal/transport/grpc/user"
	http "github.com/yi-tech/go-user-service/internal/transport/http"
	httpAuth "github.com/yi-tech/go-user-service/internal/transport/http/auth"
	httpUser "github.com/yi-tech/go-user-service/internal/transport/http/user"
)

// ProvideGRPCConfig provides the gRPC server configuration
func ProvideGRPCConfig(cfg *config.Config) *grpc.Config {
	return &grpc.Config{
		GRPCPort: cfg.GRPC.Port,
		HTTPPort: cfg.GRPC.Port + 1, // Use next port for HTTP gateway to avoid conflict with main HTTP server
	}
}

// ProvideGRPCServer creates a new gRPC server
func ProvideGRPCServer(userService serviceUser.UserService, authService domainAuth.AuthService, logger *zap.Logger, cfg *grpc.Config) *grpc.Server {
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
func ProvideUserRepository(db *gorm.DB) repoUser.UserRepository {
	return repoUser.NewUserRepository(db)
}

func ProvideAuthRepository(redis *redis.Client) domainAuth.AuthRepository {
	return repoAuth.NewAuthRepository(redis)
}

// Provider functions for services
func ProvideUserService(repo repoUser.UserRepository) serviceUser.UserService {
	return serviceUser.NewUserService(repo)
}

func ProvideAuthService(userService serviceUser.UserService, authRepo domainAuth.AuthRepository, cfg *config.Config) domainAuth.AuthService {
	return serviceAuth.NewService(userService, authRepo, cfg)
}

// Provider functions for HTTP handlers
func ProvideUserHttpHandler(userService serviceUser.UserService, logger *zap.Logger) *httpUser.Handler {
	return httpUser.NewHandler(userService, logger)
}

func ProvideAuthHttpHandler(authService domainAuth.AuthService, logger *zap.Logger) *httpAuth.Handler {
	return httpAuth.NewHandler(authService, logger)
}

// Provider functions for gRPC handlers
func ProvideUserGrpcHandler(userService serviceUser.UserService, logger *zap.Logger) *grpcUser.Handler {
	return grpcUser.NewHandler(userService, logger)
}

func ProvideAuthGrpcHandler(authService domainAuth.AuthService, logger *zap.Logger) *grpcAuth.Handler {
	return grpcAuth.NewHandler(authService, logger)
}

// Provider function for middleware
func ProvideAuthMiddleware(authService domainAuth.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return middleware.AuthMiddleware(authService, logger)
}

// Provider function for router
func ProvideRouter(userHandler *httpUser.Handler, authHandler *httpAuth.Handler, authService domainAuth.AuthService, logger *zap.Logger) *gin.Engine {
	return http.NewRouter(userHandler, authHandler, authService, logger)
}

// ProvideHTTPServer creates a new HTTP server
func ProvideHTTPServer(router *gin.Engine, cfg *config.Config) *http.Server {
	return http.NewServer(router, cfg)
}
