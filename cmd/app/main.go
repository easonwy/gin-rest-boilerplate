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

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	appwire "github.com/yi-tech/go-user-service/cmd/app/wire"

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

	// Auto-migrate the database if needed
	// This would be handled by your database migration system

	// Set up Swagger UI
	app.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start gRPC server
	log.Printf("Starting gRPC server on port %d and HTTP gateway on port %d", app.GRPCServer.Config().GRPCPort, app.GRPCServer.Config().HTTPPort)
	if err := app.GRPCServer.Start(); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}

	// Start HTTP server in a goroutine
	httpPort := app.Config.App.Port
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: app.Router,
	}

	go func() {
		log.Printf("Starting HTTP server on :%d", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	log.Printf("Application started successfully. Access Swagger UI at http://localhost:%d/swagger/index.html", app.Config.App.Port)

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a context with a timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown the main HTTP server
	log.Println("Shutting down HTTP server...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Shutdown the gRPC server
	log.Println("Shutting down gRPC server...")
	if err := app.GRPCServer.Stop(); err != nil {
		log.Printf("gRPC server shutdown error: %v", err)
	}

	log.Println("Server exiting")
}
