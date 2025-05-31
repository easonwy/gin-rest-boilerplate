package userpb

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/yi-tech/go-user-service/internal/user/dto"
	"github.com/yi-tech/go-user-service/internal/user/model"
	userservice "github.com/yi-tech/go-user-service/internal/user/service"
	userpb "github.com/yi-tech/go-user-service/api/proto/user/v1"
)

// UserServer implements the UserServiceServer interface
type UserServer struct {
	userpb.UnimplementedUserServiceServer
	userService userservice.UserService
}

// NewUserServer creates a new UserServer
func NewUserServer(userService userservice.UserService) *UserServer {
	return &UserServer{
		userService: userService,
	}
}

// Register handles user registration
func (s *UserServer) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.UserResponse, error) {
	// Convert protobuf request to DTO
	dtoReq := &dto.UserRegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		// Note: FullName is not part of the DTO, you may need to add it if needed
	}

	// Call the user service to register the user
	registeredUser, err := s.userService.RegisterUser(dtoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %v", err)
	}

	// Convert the registered user back to protobuf response
	return &userpb.UserResponse{
		User: userToProto(registeredUser),
	}, nil
}

// Login handles user login
func (s *UserServer) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	// In a real implementation, you would call your authentication service here
	// For now, we'll return an error as this needs to be implemented
	return nil, fmt.Errorf("login not implemented")
}

// GetProfile retrieves a user's profile
func (s *UserServer) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.UserResponse, error) {
	// Convert string ID to uint
	userID, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	user, err := s.userService.GetUserByID(uint(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %v", err)
	}

	return &userpb.UserResponse{
		User: userToProto(user),
	}, nil
}

// UpdateProfile updates a user's profile
func (s *UserServer) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UserResponse, error) {
	// Convert string ID to uint
	userID, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	// Create update DTO with only provided fields
	updateReq := &dto.UserUpdateRequest{}

	// Only set fields that are provided in the request
	if req.Email != "" {
		updateReq.Email = &req.Email
	}
	// Note: Username and Password are also available in the DTO if needed

	updatedUser, err := s.userService.UpdateUser(uint(userID), updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %v", err)
	}

	return &userpb.UserResponse{
		User: userToProto(updatedUser),
	}, nil
}

// DeleteUser deletes a user
func (s *UserServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	// Convert string ID to uint
	userID, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	if err := s.userService.DeleteUser(uint(userID)); err != nil {
		return nil, fmt.Errorf("failed to delete user: %v", err)
	}

	return &userpb.DeleteUserResponse{
		Success: true,
	}, nil
}

// Helper function to convert domain User to protobuf User
func userToProto(u *model.User) *userpb.User {
	if u == nil {
		return nil
	}

	userProto := &userpb.User{
		Id:       strconv.FormatUint(uint64(u.ID), 10),
		Username: u.Username,
		Email:    u.Email,
	}

	// Set timestamps if they exist
	if !u.CreatedAt.IsZero() {
		userProto.CreatedAt = timestamppb.New(u.CreatedAt)
	}

	if !u.UpdatedAt.IsZero() {
		userProto.UpdatedAt = timestamppb.New(u.UpdatedAt)
	}

	return userProto
}
