package service

import (
	"errors"
	"fmt"

	"github.com/tapas/go-user-service/internal/user/dto"
	"github.com/tapas/go-user-service/internal/user/model"
	"github.com/tapas/go-user-service/internal/user/repository"
)

var (
	ErrUserAlreadyExists = errors.New("user with this username or email already exists")
)

// UserService defines the interface for user-related business logic.
type UserService interface {
	RegisterUser(req *dto.UserRegisterRequest) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	GetUserByID(id uint) (*model.User, error)
	UpdateUser(id uint, req *dto.UserUpdateRequest) (*model.User, error)
	DeleteUser(id uint) error
	// TODO: Add other user related service methods
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new instance of UserService.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) RegisterUser(req *dto.UserRegisterRequest) (*model.User, error) {
	// Check if user already exists by username or email
	existingUserByUsername, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check user by username: %w", err)
	}
	if existingUserByUsername != nil {
		return nil, ErrUserAlreadyExists
	}

	existingUserByEmail, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user by email: %w", err)
	}
	if existingUserByEmail != nil {
		return nil, ErrUserAlreadyExists
	}

	// Create new user model
	user := &model.User{
		Username: req.Username,
		Password: req.Password, // Password will be hashed in the model method
		Email:    req.Email,
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

func (s *userService) GetUserByEmail(email string) (*model.User, error) {
	return s.userRepo.GetUserByEmail(email)
}

func (s *userService) GetUserByID(id uint) (*model.User, error) {
	return s.userRepo.GetUserByID(id)
}

func (s *userService) UpdateUser(id uint, req *dto.UserUpdateRequest) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Password != nil {
		// Hash new password
		user.Password = *req.Password
		if err := user.HashPassword(); err != nil {
			return nil, fmt.Errorf("failed to hash new password: %w", err)
		}
	}

	err = s.userRepo.UpdateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *userService) DeleteUser(id uint) error {
	// Check if user exists before deleting
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("failed to check user before deleting: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.userRepo.DeleteUser(id)
}
