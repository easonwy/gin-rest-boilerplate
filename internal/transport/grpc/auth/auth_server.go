package auth

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	authpb "github.com/yi-tech/go-user-service/api/proto/auth/v1"
	userpb "github.com/yi-tech/go-user-service/api/proto/user/v1"
	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth"
)

// AuthServer implements the AuthService gRPC service
type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	authService domainAuth.AuthService
	logger      *zap.Logger
}

// NewAuthServer creates a new AuthServer
func NewAuthServer(authService domainAuth.AuthService, logger *zap.Logger) *AuthServer {
	return &AuthServer{
		authService: authService,
		logger:      logger,
	}
}

// Login authenticates a user and returns tokens
func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.TokenResponse, error) {
	s.logger.Info("Login request received", zap.String("email", req.Email))

	// Validate input parameters
	if req.Email == "" {
		s.logger.Error("Login failed: email is required")
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		s.logger.Error("Login failed: password is required")
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}

	// Create domainAuth.LoginInput from the request
	loginInput := domainAuth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	// Call the auth service to authenticate the user
	tokenPair, err := s.authService.Login(ctx, loginInput)
	if err != nil {
		s.logger.Error("Login failed", zap.Error(err))

		// Check for specific error types
		if err.Error() == "invalid credentials" {
			return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		// For other errors (like database errors), return Internal error code
		return nil, status.Errorf(codes.Internal, "authentication failed: %v", err)
	}

	return &authpb.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthServer) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.TokenResponse, error) {
	s.logger.Info("RefreshToken request received")

	// Validate input parameters
	if req.RefreshToken == "" {
		s.logger.Error("Token refresh failed: refresh token is required")
		return nil, status.Errorf(codes.InvalidArgument, "refresh token is required")
	}

	// Call the auth service to refresh the token
	tokenPair, err := s.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Error("Token refresh failed", zap.Error(err))

		// Check for specific error types
		if err.Error() == "invalid token" || err.Error() == "session not found" {
			return nil, status.Errorf(codes.Unauthenticated, "token refresh failed: %v", err)
		}

		// For other errors (like database errors), return Internal error code
		return nil, status.Errorf(codes.Internal, "token refresh failed: %v", err)
	}

	return &authpb.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// Logout invalidates a refresh token
func (s *AuthServer) Logout(ctx context.Context, req *authpb.LogoutRequest) (*emptypb.Empty, error) {
	s.logger.Info("Logout request received")

	// Validate input parameters
	if req.RefreshToken == "" {
		s.logger.Error("Logout failed: refresh token is required")
		return nil, status.Errorf(codes.InvalidArgument, "refresh token is required")
	}

	// Extract user ID from context metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.logger.Error("Logout failed: no metadata in context")
		return nil, status.Errorf(codes.Unauthenticated, "no metadata in context")
	}

	userIDValues := md.Get("user-id")
	if len(userIDValues) == 0 {
		s.logger.Error("Logout failed: no user ID in metadata")
		return nil, status.Errorf(codes.Unauthenticated, "no user ID in metadata")
	}

	userID, err := uuid.Parse(userIDValues[0])
	if err != nil {
		s.logger.Error("Invalid user ID format", zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "invalid user ID format: %v", err)
	}

	// Call the auth service to logout the user
	err = s.authService.Logout(ctx, userID)
	if err != nil {
		// Check if it's a "session not found" error, which we'll treat as a success
		if err.Error() == "session not found" {
			s.logger.Warn("Session not found during logout, treating as success")
			return &emptypb.Empty{}, nil
		}

		s.logger.Error("Logout failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "logout failed: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// ValidateToken validates an access token
func (s *AuthServer) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	s.logger.Info("ValidateToken request received")

	// Validate input parameters
	if req.AccessToken == "" {
		s.logger.Error("Token validation failed: access token is required")
		return nil, status.Errorf(codes.InvalidArgument, "access token is required")
	}

	// Call the auth service to validate the token
	userID, err := s.authService.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		s.logger.Error("Token validation failed", zap.Error(err))
		// Check if it's an invalid token error
		if err.Error() == "invalid token" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}
		return nil, status.Errorf(codes.Internal, "token validation failed: %v", err)
	}

	return &authpb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID.String(),
	}, nil
}

// GetUserFromToken retrieves a user from an access token
func (s *AuthServer) GetUserFromToken(ctx context.Context, req *authpb.GetUserFromTokenRequest) (*userpb.User, error) {
	s.logger.Info("GetUserFromToken request received")

	// Validate input parameters
	if req.AccessToken == "" {
		s.logger.Error("GetUserFromToken failed: access token is required")
		return nil, status.Errorf(codes.InvalidArgument, "access token is required")
	}

	// First validate the token and get the user ID
	userID, err := s.authService.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		s.logger.Error("Token validation failed", zap.Error(err))
		// Check if it's an invalid token error
		if err.Error() == "invalid token" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Get the user service from the context or dependency injection
	// For now, we'll return a minimal user object with just the ID
	return &userpb.User{
		Id:        userID.String(),
		Email:     "user@example.com", // This is a placeholder
		FirstName: "User",             // This is a placeholder
		LastName:  "Name",             // This is a placeholder
		IsActive:  true,               // Assuming active user
	}, nil
}
