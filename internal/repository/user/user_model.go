package user

import (
	"time"

	"github.com/google/uuid"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
)

// UserModel represents the user structure for database interactions.
type UserModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Username  string    `gorm:"uniqueIndex;not null"`
	FirstName string
	LastName  string
	Password  string `gorm:"not null"`
	Email     string `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for the UserModel.
func (UserModel) TableName() string {
	return "users"
}

// ToDomainUser converts a UserModel to a domainUser.User.
func ToDomainUser(userModel *UserModel) *domainUser.User {
	if userModel == nil {
		return nil
	}
	return &domainUser.User{
		ID:        userModel.ID,
		Username:  userModel.Username,
		FirstName: userModel.FirstName,
		LastName:  userModel.LastName,
		Password:  userModel.Password,
		Email:     userModel.Email,
		CreatedAt: userModel.CreatedAt,
		UpdatedAt: userModel.UpdatedAt,
	}
}

// FromDomainUser converts a domainUser.User to a UserModel.
func FromDomainUser(domainUser *domainUser.User) *UserModel {
	if domainUser == nil {
		return nil
	}
	return &UserModel{
		ID:        domainUser.ID,
		Username:  domainUser.Username,
		FirstName: domainUser.FirstName,
		LastName:  domainUser.LastName,
		Password:  domainUser.Password,
		Email:     domainUser.Email,
		CreatedAt: domainUser.CreatedAt,
		UpdatedAt: domainUser.UpdatedAt,
	}
}
