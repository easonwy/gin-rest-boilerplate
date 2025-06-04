package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
	repoUser "github.com/yi-tech/go-user-service/internal/repository/user"
)

var (
	ErrUserAlreadyExists = errors.New("user with this username or email already exists")
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
	userRepo repoUser.UserRepository
}

// NewUserService creates a new instance of UserService.
func NewUserService(userRepo repoUser.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

// Register creates a new user with the provided credentials
func (s *userService) Register(ctx context.Context, email, password, firstName, lastName string) (*domainUser.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return nil, errors.New("user already exists")
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
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	return s.userRepo.GetUserByEmail(email)
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error) {
	return s.userRepo.GetUserByID(uint(id.ID()))
}

func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domainUser.User, error) {
	// Get existing user
	existingUser, err := s.userRepo.GetUserByID(uint(id.ID()))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	existingUser.FirstName = firstName
	existingUser.LastName = lastName

	// Update user
	if err := s.userRepo.UpdateUser(existingUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return existingUser, nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, params domainUser.UpdateUserParams) (*domainUser.User, error) {
	// Get existing user
	existingUser, err := s.userRepo.GetUserByID(uint(id.ID()))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if email is being changed and if it's already in use
	if params.Email != "" && params.Email != existingUser.Email {
		existingUserByEmail, err := s.userRepo.GetUserByEmail(params.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email availability: %w", err)
		}
		if existingUserByEmail != nil {
			return nil, fmt.Errorf("email already in use")
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
