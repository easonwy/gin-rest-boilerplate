package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yi-tech/go-user-service/internal/domain/user"
	userService "github.com/yi-tech/go-user-service/internal/service/user"
	pb "github.com/yi-tech/go-user-service/api/proto/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler handles gRPC requests for user operations
type Handler struct {
	pb.UnimplementedUserServiceServer
	userService userService.UserService
	logger      *zap.Logger
}

// NewHandler creates a new user gRPC handler
func NewHandler(userService userService.UserService, logger *zap.Logger) *Handler {
	return &Handler{
		userService: userService,
		logger:      logger,
	}
}

// Register handles user registration
func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.User, error) {
	// Validate request
	if req.Email == "" || req.Password == "" || req.FirstName == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing required fields")
	}

	// Register user with domain service
	newUser, err := h.userService.Register(
		ctx,
		req.Email,
		req.Password,
		req.FirstName,
		req.LastName,
	)
	if err != nil {
		// Check for specific errors
		if err.Error() == "user already exists" {
			return nil, status.Error(codes.AlreadyExists, "Email already exists")
		}
		h.logger.Error("Failed to register user", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to register user")
	}

	return toProtoUser(newUser), nil
}

// GetUserByID handles retrieving a user by ID
func (h *Handler) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.User, error) {
	// Convert string ID to UUID
	userUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	user, err := h.userService.GetByID(ctx, userUUID)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		h.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", req.Id))
		return nil, status.Error(codes.Internal, "Failed to get user")
	}

	return toProtoUser(user), nil
}

// GetUserByEmail handles retrieving a user by email
func (h *Handler) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.User, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}

	user, err := h.userService.GetByEmail(ctx, req.Email)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		h.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", req.Email))
		return nil, status.Error(codes.Internal, "Failed to get user")
	}

	return toProtoUser(user), nil
}

// UpdateUser handles updating a user's details
func (h *Handler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	// Convert string ID to UUID
	userUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	if req.FirstName == "" {
		return nil, status.Error(codes.InvalidArgument, "First name is required")
	}

	// Update user with domain service
	updatedUser, err := h.userService.UpdateUser(ctx, userUUID, req.FirstName, req.LastName)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		h.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", req.Id))
		return nil, status.Error(codes.Internal, "Failed to update user")
	}

	return toProtoUser(updatedUser), nil
}

// UpdatePassword handles updating a user's password
func (h *Handler) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*emptypb.Empty, error) {
	// Convert string ID to UUID
	userUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "Current password and new password are required")
	}

	// Update password with domain service
	err = h.userService.UpdatePassword(ctx, userUUID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		h.logger.Error("Failed to update password", zap.Error(err), zap.String("user_id", req.Id))
		return nil, status.Error(codes.Internal, "Failed to update password")
	}

	return &emptypb.Empty{}, nil
}

// DeleteUser handles deleting a user
func (h *Handler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	// Convert string ID to UUID
	userUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	err = h.userService.DeleteUser(ctx, userUUID)
	if err != nil {
		// Generic error handling for now
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		h.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", req.Id))
		return nil, status.Error(codes.Internal, "Failed to delete user")
	}

	return &emptypb.Empty{}, nil
}

// toProtoUser converts a domain user model to a protobuf user message
func toProtoUser(user *user.User) *pb.User {
	return &pb.User{
		Id:        fmt.Sprintf("%d", user.ID),
		Email:     user.Email,
		FirstName: user.Username, // Using username as first name
		LastName:  "",            // No last name in the domain model
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}
