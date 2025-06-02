package user

import (
	"context"

	"github.com/google/uuid"
)

// UserService defines the interface for user business logic
type UserService interface {
	// Register creates a new user
	Register(ctx context.Context, email, password, firstName, lastName string) (*User, error)
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)
	
	// UpdateUser updates user details
	UpdateUser(ctx context.Context, id uuid.UUID, firstName, lastName string) (*User, error)
	
	// UpdatePassword changes a user's password
	UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error
	
	// DeleteUser removes a user
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
