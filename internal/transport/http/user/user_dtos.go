package user

import (
	"encoding/json"
	"time"
)

// UserRegisterRequest defines the request body for user registration.
type UserRegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
}

// UserResponse defines the common response structure for a user.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName,omitempty"`
	LastName  string    `json:"lastName,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// MarshalJSON implements custom JSON marshaling for UserResponse to ensure consistent timestamp format
func (u UserResponse) MarshalJSON() ([]byte, error) {
	type Alias UserResponse
	return json.Marshal(&struct {
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		*Alias
	}{
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		Alias:     (*Alias)(&u),
	})
}

// UserUpdateRequest defines the request body for updating user profile information.
type UserUpdateRequest struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     *string `json:"email" binding:"omitempty,email"`
}

// UpdatePasswordRequest defines the request body for updating a user's password.
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

// UpdateCurrentUserProfileRequest defines the request body for updating the current user's profile.
type UpdateCurrentUserProfileRequest struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     *string `json:"email" binding:"omitempty,email"`
}
