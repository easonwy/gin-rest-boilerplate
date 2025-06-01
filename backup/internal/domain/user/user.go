package user

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username" binding:"required"`
	Password  string    `gorm:"not null" json:"-" binding:"required"` // Store hashed password, exclude from JSON output
	Email     string    `gorm:"uniqueIndex;not null" json:"email" binding:"required,email"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for the User model.
func (User) TableName() string {
	return "users"
}

// HashPassword hashes the user's password.
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword checks if the provided password matches the hashed password.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
