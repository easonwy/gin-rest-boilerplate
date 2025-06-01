package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/yi-tech/go-user-service/internal/domain/user"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailExists      = errors.New("email already exists")
	ErrInvalidPassword  = errors.New("invalid password")
)

// userService implements the domain user.Service interface
type userService struct {
	repo   user.Repository
	logger *zap.Logger
}

// NewUserService creates a new user service
func NewUserService(repo user.Repository, logger *zap.Logger) user.Service {
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

// Register creates a new user
func (s *userService) Register(ctx context.Context, email, password, firstName, lastName string) (*user.User, error) {
	// Check if user already exists
	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailExists
	}

	// Create new user
	newUser, err := user.NewUser(email, password, firstName, lastName)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Save user to repository
	if err := s.repo.Create(ctx, newUser); err != nil {
		s.logger.Error("Failed to save user", zap.Error(err))
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return newUser, nil
}

// GetByID retrieves a user by ID
func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", id.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	if user == nil {
		return nil, ErrUserNotFound
	}
	
	return user, nil
}

// GetByEmail retrieves a user by email
func (s *userService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	if user == nil {
		return nil, ErrUserNotFound
	}
	
	return user, nil
}

// UpdateUser updates user details
func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, firstName, lastName string) (*user.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user for update", zap.Error(err), zap.String("user_id", id.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	if user == nil {
		return nil, ErrUserNotFound
	}
	
	// Update user details
	user.UpdateDetails(firstName, lastName)
	
	// Save updated user
	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", id.String()))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	return user, nil
}

// UpdatePassword changes a user's password
func (s *userService) UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user for password update", zap.Error(err), zap.String("user_id", id.String()))
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	if user == nil {
		return ErrUserNotFound
	}
	
	// Validate current password
	if !user.ValidatePassword(currentPassword) {
		return ErrInvalidPassword
	}
	
	// Update password
	if err := user.UpdatePassword(newPassword); err != nil {
		s.logger.Error("Failed to hash new password", zap.Error(err))
		return fmt.Errorf("failed to update password: %w", err)
	}
	
	// Save updated user
	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to save updated password", zap.Error(err), zap.String("user_id", id.String()))
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	return nil
}

// DeleteUser removes a user
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Check if user exists
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user for deletion", zap.Error(err), zap.String("user_id", id.String()))
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	if user == nil {
		return ErrUserNotFound
	}
	
	// Delete user
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", id.String()))
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	return nil
}
