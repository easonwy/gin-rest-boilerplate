package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yi-tech/go-user-service/internal/config"
)

// Server represents the HTTP server
type Server struct {
	router *gin.Engine
	server *http.Server
	cfg    *config.Config
}

// NewServer creates a new HTTP server
func NewServer(router *gin.Engine, cfg *config.Config) *Server {
	return &Server{
		router: router,
		cfg:    cfg,
	}
}

// Router returns the Gin router
func (s *Server) Router() *gin.Engine {
	return s.router
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.App.Port),
		Handler: s.router,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	return s.server.Shutdown(ctx)
}

// WithMiddleware adds middleware to the router
func (s *Server) WithMiddleware(middleware ...gin.HandlerFunc) {
	s.router.Use(middleware...)
}

// WithTimeout sets the read/write timeout for the server
func (s *Server) WithTimeout(read, write time.Duration) {
	if s.server != nil {
		s.server.ReadTimeout = read
		s.server.WriteTimeout = write
	}
}
