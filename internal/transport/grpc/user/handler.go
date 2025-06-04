package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	userpb "github.com/yi-tech/go-user-service/api/proto/user/v1"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
	serviceUser "github.com/yi-tech/go-user-service/internal/service/user"
)

// Handler is a wrapper for the UserServer to match the wire.go expectations
type Handler struct {
	*UserServer
}

// NewHandler creates a new user gRPC handler
func NewHandler(userService serviceUser.UserService, logger *zap.Logger) *Handler {
	return &Handler{
		UserServer: NewUserServer(userService, logger),
	}
}

// GetServer returns the underlying UserServer for registration with gRPC
func (h *Handler) GetServer() userpb.UserServiceServer {
	return h.UserServer
}

// Register handles the Register gRPC request
// This is a wrapper around Register to maintain compatibility with tests
func (h *Handler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.User, error) {
	// Validate required fields
	if req.Email == "" || req.Password == "" || req.FirstName == "" {
		return nil, status.Error(codes.InvalidArgument, "Email, password, and first name are required")
	}

	// Populate RegisterUserInput
	userInput := serviceUser.RegisterUserInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Call the user service to register the user
	user, err := h.userService.Register(ctx, userInput)
	if err != nil {
		// Check for specific error types
		if errors.Is(err, errors.New("user already exists")) || err.Error() == "user already exists" {
			return nil, status.Error(codes.AlreadyExists, "User already exists")
		}

		h.logger.Error("User registration failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "user registration failed: %v", err)
	}

	// Convert domain user to protobuf user
	return h.userToPb(user), nil
}

// GetUserByID handles the GetUserByID gRPC request
// This is a wrapper around GetProfile to maintain compatibility with tests
func (h *Handler) GetUserByID(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.User, error) {
	// Validate UUID
	userID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	// Get user from service
	user, err := h.userService.GetByID(ctx, userID)
	if err != nil {
		// Check for specific error types
		if errors.Is(err, errors.New("user not found")) || err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}

		// Log the error
		h.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", req.GetId()))
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	// Convert domain user to protobuf user
	return h.userToPb(user), nil
}

// GetUserByEmailRequest is a custom type for the test
type GetUserByEmailRequest struct {
	Email string
}

// GetId returns the email
func (r *GetUserByEmailRequest) GetEmail() string {
	return r.Email
}

// GetUserByEmail handles the GetUserByEmail gRPC request
// This is a custom implementation for test compatibility
func (h *Handler) GetUserByEmail(ctx context.Context, req *GetUserByEmailRequest) (*userpb.User, error) {
	// Validate email
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}

	// Get user from service
	user, err := h.userService.GetByEmail(ctx, req.GetEmail())
	if err != nil {
		// Check for specific error types
		if errors.Is(err, errors.New("user not found")) || err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}

		// Log the error
		h.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", req.GetEmail()))
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	// Convert domain user to protobuf user
	return h.userToPb(user), nil
}

// UpdateUserRequest is a custom type for the test
type UpdateUserRequest struct {
	Id        string
	FirstName string
	LastName  string
}

// GetId returns the ID
func (r *UpdateUserRequest) GetId() string {
	return r.Id
}

// GetFirstName returns the first name
func (r *UpdateUserRequest) GetFirstName() string {
	return r.FirstName
}

// GetLastName returns the last name
func (r *UpdateUserRequest) GetLastName() string {
	return r.LastName
}

// UpdateUser handles the UpdateUser gRPC request
// This is a custom implementation for test compatibility
func (h *Handler) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*userpb.User, error) {
	// Validate UUID
	userID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	// Validate required fields
	if req.GetFirstName() == "" {
		return nil, status.Error(codes.InvalidArgument, "First name is required")
	}

	// Create UpdateUserParams
	params := domainUser.UpdateUserParams{
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	}

	// Update user in service
	user, err := h.userService.Update(ctx, userID, params)
	if err != nil {
		// Check for specific error types
		if errors.Is(err, errors.New("user not found")) || err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}

		// Log the error
		h.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", req.GetId()))
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	// Convert domain user to protobuf user
	return h.userToPb(user), nil
}

// UpdatePasswordRequest is a custom type for the test
type UpdatePasswordRequest struct {
	Id              string
	CurrentPassword string
	NewPassword     string
}

// GetId returns the ID
func (r *UpdatePasswordRequest) GetId() string {
	return r.Id
}

// GetCurrentPassword returns the current password
func (r *UpdatePasswordRequest) GetCurrentPassword() string {
	return r.CurrentPassword
}

// GetNewPassword returns the new password
func (r *UpdatePasswordRequest) GetNewPassword() string {
	return r.NewPassword
}

// UpdatePassword handles the UpdatePassword gRPC request
// This is a custom implementation for test compatibility
func (h *Handler) UpdatePassword(ctx context.Context, req *UpdatePasswordRequest) (*emptypb.Empty, error) {
	// Validate UUID
	userID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	// Validate required fields
	if req.GetCurrentPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Current password is required")
	}

	if req.GetNewPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "New password is required")
	}

	// Update password in service
	err = h.userService.UpdatePassword(ctx, userID, req.GetCurrentPassword(), req.GetNewPassword())
	if err != nil {
		// Check for specific error types
		if errors.Is(err, errors.New("user not found")) || err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}

		if errors.Is(err, errors.New("invalid current password")) || err.Error() == "invalid current password" {
			return nil, status.Error(codes.InvalidArgument, "Invalid current password")
		}

		// Log the error
		h.logger.Error("Failed to update password", zap.Error(err), zap.String("user_id", req.GetId()))
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	return &emptypb.Empty{}, nil
}

// DeleteUser handles the DeleteUser gRPC request
// This is a custom implementation for test compatibility
func (h *Handler) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*emptypb.Empty, error) {
	// Validate UUID
	userID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	// Delete user in service
	err = h.userService.DeleteUser(ctx, userID)
	if err != nil {
		// Check for specific error types
		if errors.Is(err, errors.New("user not found")) || err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}

		// Log the error
		h.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", req.GetId()))
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	return &emptypb.Empty{}, nil
}

// userToPb converts a domain user to a protobuf user
func (h *Handler) userToPb(user *domainUser.User) *userpb.User {
	var createdAt, updatedAt *timestamppb.Timestamp

	if !user.CreatedAt.IsZero() {
		createdAt = timestamppb.New(user.CreatedAt)
	}

	if !user.UpdatedAt.IsZero() {
		updatedAt = timestamppb.New(user.UpdatedAt)
	}

	return &userpb.User{
		Id:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  true, // Assuming all users are active by default
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
