package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
	"gorm.io/gorm"
)

// UserService defines the interface for user-related business logic.
type UserService interface {
	// Register creates a new user
	Register(ctx context.Context, email, password, firstName, lastName string) (*domainUser.User, error)

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*domainUser.User, error)

	// UpdateUser updates user details
	UpdateUser(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domainUser.User, error)

	// Update updates user details with the provided parameters
	Update(ctx context.Context, id uuid.UUID, params domainUser.UpdateUserParams) (*domainUser.User, error)

	// UpdatePassword changes a user's password
	UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error

	// DeleteUser removes a user
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	userRepo domainUser.Repository
}

// NewUserService creates a new instance of UserService.
func NewUserService(userRepo domainUser.Repository) UserService {
	return &userService{userRepo: userRepo}
}

// Register creates a new user with the provided credentials
func (s *userService) Register(ctx context.Context, email, password, firstName, lastName string) (*domainUser.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// If GORM's record not found, it's not an error for this check, means email is available
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing user: %w", err)
		}
	}

	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Create new user
	user := &domainUser.User{
		ID:        uuid.New(),
		Username:  email, // Set username to email to satisfy the not-null constraint
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Hash password
	if err := user.HashPassword(); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Assuming repo returns gorm.ErrRecordNotFound which should be translated
		// For now, let's expect direct error or nil user from repo for not found
		return nil, fmt.Errorf("failed to get user by email from repository: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		// Assuming repo returns gorm.ErrRecordNotFound which should be translated
		// For now, let's expect direct error or nil user from repo for not found
		return nil, fmt.Errorf("failed to get user by id from repository: %w", err)
	}
	if user == nil { // This check is key if repo returns (nil, nil) on not found
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domainUser.User, error) {
	// Get existing user
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for update: %w", err)
	}
	if existingUser == nil {
		return nil, ErrUserNotFound
	}

	// Update fields
	existingUser.FirstName = firstName
	existingUser.LastName = lastName

	// Update user
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return existingUser, nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, params domainUser.UpdateUserParams) (*domainUser.User, error) {
	// Get existing user
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for update: %w", err)
	}
	if existingUser == nil {
		return nil, ErrUserNotFound
	}

	// Check if email is being changed and if it's already in use
	if params.Email != "" && params.Email != existingUser.Email {
		// Need to handle potential errors from GetByEmail itself
		conflictingUser, err := s.userRepo.GetByEmail(ctx, params.Email)
		if err != nil {
			// If GORM's record not found, it's not an error for this check, means email is available for use by current user
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("failed to check email availability: %w", err)
			}
		}
		if conflictingUser != nil {
			return nil, ErrEmailInUse
		}
		existingUser.Email = params.Email
	}

	// Update other fields if provided
	if params.FirstName != "" {
		existingUser.FirstName = params.FirstName
	}

	if params.LastName != "" {
		existingUser.LastName = params.LastName
	}

	// Update user
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return existingUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Get existing user
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user for delete: %w", err)
	}
	if existingUser == nil {
		return ErrUserNotFound
	}

	// Delete user
	return s.userRepo.Delete(ctx, id)
}

func (s *userService) UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error {
	// Get existing user
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user for password update: %w", err)
	}
	if existingUser == nil {
		return ErrUserNotFound
	}

	// Verify current password
	if !existingUser.CheckPassword(currentPassword) {
		return ErrIncorrectPassword
	}

	// Update password
	existingUser.Password = newPassword
	if err := existingUser.HashPassword(); err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Save user
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
