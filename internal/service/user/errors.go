package user

import "errors"

// Service-level errors for user operations
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailInUse        = errors.New("email already in use")
	ErrIncorrectPassword = errors.New("incorrect current password")
	ErrUserAlreadyExists = errors.New("user already exists") // Moved from user_service.go
)
