package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yi-tech/go-user-service/internal/domain/auth"
	"github.com/yi-tech/go-user-service/internal/domain/user"
	pb "github.com/yi-tech/go-user-service/api/proto/auth"
	userPb "github.com/yi-tech/go-user-service/api/proto/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler handles gRPC requests for authentication operations
type Handler struct {
	pb.UnimplementedAuthServiceServer
	authService auth.AuthService
	logger      *zap.Logger
}

// NewHandler creates a new authentication gRPC handler
func NewHandler(authService auth.AuthService, logger *zap.Logger) *Handler {
	return &Handler{
		authService: authService,
		logger:      logger,
	}
}

// Login handles user login
func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.TokenResponse, error) {
	// Validate request
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "Email and password are required")
	}

	// Authenticate user
	tokenPair, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "invalid credentials" {
			return nil, status.Error(codes.Unauthenticated, "Invalid email or password")
		}
		h.logger.Error("Failed to login user", zap.Error(err), zap.String("email", req.Email))
		return nil, status.Error(codes.Internal, "Failed to login")
	}

	return &pb.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.TokenResponse, error) {
	// Validate request
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "Refresh token is required")
	}

	// Refresh token
	tokenPair, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "invalid token" || err.Error() == "session not found" {
			return nil, status.Error(codes.Unauthenticated, "Invalid or expired refresh token")
		}
		h.logger.Error("Failed to refresh token", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to refresh token")
	}

	return &pb.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// Logout handles user logout
func (h *Handler) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "Refresh token is required")
	}

	// Get user ID from context or metadata
	userID, err := h.getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}

	// Logout user
	err = h.authService.Logout(ctx, userID)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "invalid token" || err.Error() == "session not found" {
			// Still return success even if token was invalid or session not found
			return &emptypb.Empty{}, nil
		}
		h.logger.Error("Failed to logout user", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to logout")
	}

	return &emptypb.Empty{}, nil
}

// ValidateToken handles token validation
func (h *Handler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "Access token is required")
	}

	// Validate token
	userID, err := h.authService.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "invalid token" {
			return nil, status.Error(codes.Unauthenticated, "Invalid or expired token")
		}
		h.logger.Error("Failed to validate token", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to validate token")
	}

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID.String(), // Convert UUID to string
	}, nil
}

// GetUserFromToken retrieves user information from a token
func (h *Handler) GetUserFromToken(ctx context.Context, req *pb.GetUserFromTokenRequest) (*userPb.User, error) {
	// This method is not supported by the current auth service interface
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// getUserIDFromContext extracts user ID from the context
func (h *Handler) getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, fmt.Errorf("no metadata in context")
	}

	userIDs := md.Get("user-id")
	if len(userIDs) == 0 {
		return uuid.Nil, fmt.Errorf("no user ID in metadata")
	}

	// Parse string ID to uuid.UUID
	userID, err := uuid.Parse(userIDs[0])
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	return userID, nil
}

// toProtoUser converts a domain user model to a protobuf user message
func toProtoUser(user *user.User) *userPb.User {
	return &userPb.User{
		Id:        fmt.Sprintf("%d", user.ID), // Convert uint to string
		Email:     user.Email,
		FirstName: "", // Domain model doesn't have FirstName, using empty string
		LastName:  "", // Domain model doesn't have LastName, using empty string
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}
