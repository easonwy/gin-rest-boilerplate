package user

import (
	"github.com/yi-tech/go-user-service/internal/domain/user"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	CreateUser(user *user.User) error
	GetUserByUsername(username string) (*user.User, error)
	GetUserByEmail(email string) (*user.User, error)
	GetUserByID(id uint) (*user.User, error)
	UpdateUser(user *user.User) error
	DeleteUser(id uint) error
	// TODO: Add other user related repository methods
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *user.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetUserByUsername(username string) (*user.User, error) {
	var user user.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByEmail(email string) (*user.User, error) {
	var user user.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(id uint) (*user.User, error) {
	var user user.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateUser(user *user.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) DeleteUser(id uint) error {
	return r.db.Delete(&user.User{}, id).Error
}
