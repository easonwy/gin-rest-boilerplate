package user

import (
	"context"

	"github.com/google/uuid"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of domainUser.Repository.
func NewUserRepository(db *gorm.DB) domainUser.Repository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domainUser.User) error {
	userModel := FromDomainUser(user)
	return r.db.WithContext(ctx).Create(userModel).Error
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	var userModel UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&userModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found
		}
		return nil, err
	}
	return ToDomainUser(&userModel), nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error) {
	var userModel UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&userModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found
		}
		return nil, err
	}
	return ToDomainUser(&userModel), nil
}

func (r *userRepository) Update(ctx context.Context, user *domainUser.User) error {
	userModel := FromDomainUser(user)
	return r.db.WithContext(ctx).Save(userModel).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&UserModel{}).Error
}
