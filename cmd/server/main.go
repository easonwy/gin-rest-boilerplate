package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	appwire "github.com/yi-tech/go-user-service/cmd/server/wire"

	// Import for swagger docs
	_ "github.com/yi-tech/go-user-service/docs"
)

// @title User Service API
// @version 1.0
// @description This is a sample user service server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Initialize the application
	app, err := appwire.InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// Set up Swagger UI
	app.HTTPServer.Router().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create error channel to capture server errors
	errChan := make(chan error, 2)

	// Start gRPC server in a goroutine
	go func() {
		app.Logger.Info("Starting gRPC server", 
			zap.Int("grpcPort", app.Config.GRPC.Port),
			zap.Int("grpcGatewayPort", app.Config.GRPC.Port+1))
			
		if err := app.GRPCServer.Start(); err != nil {
			app.Logger.Error("Failed to start gRPC server", zap.Error(err))
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Start HTTP server in a goroutine
	go func() {
		httpPort := app.Config.App.Port
		app.Logger.Info("Starting HTTP server", zap.Int("port", httpPort))
		
		if err := app.HTTPServer.Start(); err != nil && err != http.ErrServerClosed {
			app.Logger.Error("Failed to start HTTP server", zap.Error(err))
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	app.Logger.Info("Application started successfully", 
		zap.String("swagger", fmt.Sprintf("http://localhost:%d/swagger/index.html", app.Config.App.Port)))

	// Wait for interrupt signal or error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		app.Logger.Error("Server error", zap.Error(err))
	case sig := <-quit:
		app.Logger.Info("Received signal", zap.String("signal", sig.String()))
	}

	app.Logger.Info("Shutting down servers...")

	// Create a context with a timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown the HTTP server
	app.Logger.Info("Shutting down HTTP server...")
	if err := app.HTTPServer.Shutdown(shutdownCtx); err != nil {
		app.Logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	// Shutdown the gRPC server
	app.Logger.Info("Shutting down gRPC server...")
	if err := app.GRPCServer.Stop(); err != nil {
		app.Logger.Error("gRPC server shutdown error", zap.Error(err))
	}

	app.Logger.Info("Server exiting")
}
