package user

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	userpb "github.com/yi-tech/go-user-service/api/proto/user/v1"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
	serviceUser "github.com/yi-tech/go-user-service/internal/service/user"
)

// UserServer implements the UserService gRPC service
type UserServer struct {
	userpb.UnimplementedUserServiceServer
	userService serviceUser.UserService
	logger      *zap.Logger
}

// NewUserServer creates a new UserServer
func NewUserServer(userService serviceUser.UserService, logger *zap.Logger) *UserServer {
	return &UserServer{
		userService: userService,
		logger:      logger,
	}
}

// Register registers a new user
func (s *UserServer) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.UserResponse, error) {
	s.logger.Info("Register request received", zap.String("email", req.Email))

	// Populate RegisterUserInput
	userInput := serviceUser.RegisterUserInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Call the user service to register the user
	user, err := s.userService.Register(ctx, userInput)
	if err != nil {
		s.logger.Error("User registration failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "user registration failed: %v", err)
	}

	return s.userToResponse(user), nil
}

// Login authenticates a user
func (s *UserServer) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	s.logger.Info("Login request received", zap.String("email", req.Email))

	// Get the user by email
	user, err := s.userService.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("User not found", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		s.logger.Error("Invalid password")
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	// In a real implementation, you would generate tokens here
	// For now, we'll just return placeholders
	accessToken := "placeholder-access-token"
	refreshToken := "placeholder-refresh-token"

	return &userpb.LoginResponse{
		User:         s.userToPb(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetProfile retrieves a user profile
func (s *UserServer) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.UserResponse, error) {
	s.logger.Info("GetProfile request received", zap.String("id", req.Id))

	// Parse the ID string to UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Error("Invalid user ID format", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	// Call the user service to get the user profile
	user, err := s.userService.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Get user profile failed", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return s.userToResponse(user), nil
}

// UpdateProfile updates a user profile
func (s *UserServer) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UserResponse, error) {
	s.logger.Info("UpdateProfile request received", zap.String("id", req.Id))

	// Parse the ID string to UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Error("Invalid user ID format", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	// Create UpdateUserParams
	params := domainUser.UpdateUserParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		// Email field remains empty as per requirement, it's not part of UpdateProfileRequest
	}

	// Call the user service to update the user profile
	user, err := s.userService.Update(ctx, id, params)
	if err != nil {
		s.logger.Error("Update user profile failed", zap.Error(err))
		// TODO: Handle specific errors like ErrUserNotFound or ErrEmailInUse if necessary
		return nil, status.Errorf(codes.Internal, "update user profile failed: %v", err)
	}

	return s.userToResponse(user), nil
}

// DeleteUser deletes a user
func (s *UserServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	s.logger.Info("DeleteUser request received", zap.String("id", req.Id))

	// Parse the ID string to UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Error("Invalid user ID format", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	// Call the user service to delete the user
	err = s.userService.DeleteUser(ctx, id)
	if err != nil {
		s.logger.Error("Delete user failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "delete user failed: %v", err)
	}

	return &userpb.DeleteUserResponse{
		Success: true,
	}, nil
}

// userToResponse converts a domain user to a user response
func (s *UserServer) userToResponse(user *domainUser.User) *userpb.UserResponse {
	return &userpb.UserResponse{
		User: s.userToPb(user),
	}
}

// userToPb converts a domain user to a protobuf user
func (s *UserServer) userToPb(user *domainUser.User) *userpb.User {
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
