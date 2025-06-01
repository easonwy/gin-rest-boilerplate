package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yi-tech/go-user-service/internal/domain/user"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserModel is the GORM model for users
type UserModel struct {
	gorm.Model
	ID        string `gorm:"type:uuid;primary_key"`
	Email     string `gorm:"uniqueIndex;not null"`
	Password  string `gorm:"not null"`
	FirstName string
	LastName  string
}

// TableName sets the table name for the user model
func (UserModel) TableName() string {
	return "users"
}

// toDomainUser converts a UserModel to a domain User
func (u *UserModel) toDomainUser() *user.User {
	id, _ := uuid.Parse(u.ID)
	return &user.User{
		ID:        id,
		Email:     u.Email,
		Password:  u.Password,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// fromDomainUser converts a domain User to a UserModel
func fromDomainUser(u *user.User) *UserModel {
	return &UserModel{
		ID:        u.ID.String(),
		Email:     u.Email,
		Password:  u.Password,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}

// userRepository implements the domain user.Repository interface
type userRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, logger *zap.Logger) user.Repository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// Create stores a new user
func (r *userRepository) Create(ctx context.Context, user *user.User) error {
	userModel := fromDomainUser(user)
	
	result := r.db.WithContext(ctx).Create(userModel)
	if result.Error != nil {
		r.logger.Error("Failed to create user in database", zap.Error(result.Error))
		return result.Error
	}
	
	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var userModel UserModel
	
	result := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&userModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("Failed to get user by ID", zap.Error(result.Error), zap.String("user_id", id.String()))
		return nil, result.Error
	}
	
	return userModel.toDomainUser(), nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var userModel UserModel
	
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&userModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("Failed to get user by email", zap.Error(result.Error), zap.String("email", email))
		return nil, result.Error
	}
	
	return userModel.toDomainUser(), nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *user.User) error {
	userModel := fromDomainUser(user)
	
	result := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", user.ID.String()).Updates(userModel)
	if result.Error != nil {
		r.logger.Error("Failed to update user", zap.Error(result.Error), zap.String("user_id", user.ID.String()))
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		r.logger.Warn("No rows affected when updating user", zap.String("user_id", user.ID.String()))
		return errors.New("user not found")
	}
	
	return nil
}

// Delete removes a user by ID
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id.String())
	if result.Error != nil {
		r.logger.Error("Failed to delete user", zap.Error(result.Error), zap.String("user_id", id.String()))
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		r.logger.Warn("No rows affected when deleting user", zap.String("user_id", id.String()))
		return errors.New("user not found")
	}
	
	return nil
}
