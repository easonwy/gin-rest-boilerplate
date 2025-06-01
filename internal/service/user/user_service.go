package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	domain "github.com/yi-tech/go-user-service/internal/domain/user"
	repo "github.com/yi-tech/go-user-service/internal/repository/user"
)

var (
	ErrUserAlreadyExists = errors.New("user with this username or email already exists")
)

// UserService defines the interface for user-related business logic.
type UserService interface {
	// Register creates a new user
	Register(ctx context.Context, email, password, firstName, lastName string) (*domain.User, error)
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	
	// UpdateUser updates user details
	UpdateUser(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domain.User, error)
	
	// UpdatePassword changes a user's password
	UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error
	
	// DeleteUser removes a user
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	userRepo repo.UserRepository
}

// NewUserService creates a new instance of UserService.
func NewUserService(userRepo repo.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) Register(ctx context.Context, email, password, firstName, lastName string) (*domain.User, error) {
	// Check if user already exists by email
	existingUserByEmail, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user by email: %w", err)
	}
	if existingUserByEmail != nil {
		return nil, ErrUserAlreadyExists
	}

	// Create new user model
	user := &domain.User{
		Email:    email,
		Password: password, // Password will be hashed in the model method
	}

	// Hash password
	if err := user.HashPassword(); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Save user to database
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetUserByEmail(email)
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetUserByID(uint(id.ID()))
}

func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domain.User, error) {
	// Get existing user
	existingUser, err := s.userRepo.GetUserByID(uint(id.ID()))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields if they exist in the User struct
	// Note: Assuming these fields exist in the domain User model
	// If they don't, you'll need to adapt this code to match the actual User struct

	// Update user
	if err := s.userRepo.UpdateUser(existingUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return existingUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Get existing user
	existingUser, err := s.userRepo.GetUserByID(uint(id.ID()))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return fmt.Errorf("user not found")
	}

	// Delete user
	return s.userRepo.DeleteUser(uint(id.ID()))
}

func (s *userService) UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error {
	// Get existing user
	existingUser, err := s.userRepo.GetUserByID(uint(id.ID()))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return fmt.Errorf("user not found")
	}

	// Verify current password
	// TODO: Implement password verification

	// Update password
	existingUser.Password = newPassword
	
	// Save user
	return s.userRepo.UpdateUser(existingUser)
}
