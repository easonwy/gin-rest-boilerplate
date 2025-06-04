package user

import (
	"context"
	"errors" // Added for errors.New
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt" // Added for bcrypt in TestUpdatePassword
	"gorm.io/gorm"               // For gorm.ErrRecordNotFound

	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
)

// MockUserRepository is a mock implementation of the domainUser.Repository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domainUser.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domainUser.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Helper to create a new user for testing
func newTestUser(email, password, firstName, lastName string) *domainUser.User {
	return &domainUser.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  password, // Raw password for testing, will be hashed by service
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestRegister(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)
	ctx := context.Background()

	testUser := newTestUser("test@example.com", "password123", "Test", "User")

	t.Run("Success", func(t *testing.T) {
		// Mock GetByEmail to indicate user does not exist
		mockRepo.On("GetByEmail", ctx, testUser.Email).Return(nil, gorm.ErrRecordNotFound).Once()
		// Mock Create to succeed
		mockRepo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(nil).Once()

		createdUser, err := userService.Register(ctx, testUser.Email, testUser.Password, testUser.FirstName, testUser.LastName)

		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		assert.Equal(t, testUser.Email, createdUser.Email)
		assert.NotEmpty(t, createdUser.Password) // Password should be hashed
		assert.NotEqual(t, "password123", createdUser.Password)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Already Exists", func(t *testing.T) {
		existingUser := newTestUser("exists@example.com", "password123", "Existing", "User")
		mockRepo.On("GetByEmail", ctx, existingUser.Email).Return(existingUser, nil).Once() // User found

		createdUser, err := userService.Register(ctx, existingUser.Email, "newpass", "New", "User")

		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.EqualError(t, err, ErrUserAlreadyExists.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error on GetByEmail", func(t *testing.T) {
		mockRepo.On("GetByEmail", ctx, "error@example.com").Return(nil, errors.New("db error on get")).Once()

		createdUser, err := userService.Register(ctx, "error@example.com", "password", "Error", "User")

		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.Contains(t, err.Error(), "failed to check existing user")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error on Create", func(t *testing.T) {
		mockRepo.On("GetByEmail", ctx, "createfail@example.com").Return(nil, gorm.ErrRecordNotFound).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(errors.New("db error on create")).Once()

		createdUser, err := userService.Register(ctx, "createfail@example.com", "password", "CreateFail", "User")

		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.Contains(t, err.Error(), "failed to create user")
		mockRepo.AssertExpectations(t)
	})

	// Test for password hashing error is hard to induce reliably without direct control over bcrypt or OS resources.
}

func TestGetByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)
	ctx := context.Background()

	testUserID := uuid.New()
	testUser := newTestUser("getbyid@example.com", "password", "Get", "ByID")
	testUser.ID = testUserID

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, testUserID).Return(testUser, nil).Once()

		foundUser, err := userService.GetByID(ctx, testUserID)

		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, testUserID, foundUser.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found - Repo Returns Nil, Nil", func(t *testing.T) {
		notFoundID := uuid.New()
		mockRepo.On("GetByID", ctx, notFoundID).Return(nil, nil).Once()

		foundUser, err := userService.GetByID(ctx, notFoundID)

		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.EqualError(t, err, ErrUserNotFound.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found - Repo Returns Error (gorm.ErrRecordNotFound)", func(t *testing.T) {
		notFoundID := uuid.New()
		mockRepo.On("GetByID", ctx, notFoundID).Return(nil, gorm.ErrRecordNotFound).Once()

		foundUser, err := userService.GetByID(ctx, notFoundID)
		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.Contains(t, err.Error(), "failed to get user by id from repository")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		errorID := uuid.New()
		mockRepo.On("GetByID", ctx, errorID).Return(nil, errors.New("db error")).Once()

		foundUser, err := userService.GetByID(ctx, errorID)

		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.Contains(t, err.Error(), "db error")
		mockRepo.AssertExpectations(t)
	})
}

func TestGetByEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)
	ctx := context.Background()

	testUserEmail := "getbyemail@example.com"
	testUser := newTestUser(testUserEmail, "password", "Get", "ByEmail")

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetByEmail", ctx, testUserEmail).Return(testUser, nil).Once()

		foundUser, err := userService.GetByEmail(ctx, testUserEmail)

		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, testUserEmail, foundUser.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found - Repo Returns Nil, Nil", func(t *testing.T) {
		notFoundEmail := "notfound@example.com"
		mockRepo.On("GetByEmail", ctx, notFoundEmail).Return(nil, nil).Once()

		foundUser, err := userService.GetByEmail(ctx, notFoundEmail)

		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.EqualError(t, err, ErrUserNotFound.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found - Repo Returns Error (gorm.ErrRecordNotFound)", func(t *testing.T) {
		notFoundEmail := "notfoundgorm@example.com"
		mockRepo.On("GetByEmail", ctx, notFoundEmail).Return(nil, gorm.ErrRecordNotFound).Once()

		foundUser, err := userService.GetByEmail(ctx, notFoundEmail)
		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.Contains(t, err.Error(), "failed to get user by email from repository")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		errorEmail := "error@example.com"
		mockRepo.On("GetByEmail", ctx, errorEmail).Return(nil, errors.New("db error")).Once()

		foundUser, err := userService.GetByEmail(ctx, errorEmail)

		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.Contains(t, err.Error(), "db error")
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdate(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)
	ctx := context.Background()

	originalUserID := uuid.New()
	originalUser := &domainUser.User{
		ID: originalUserID, Email: "original@example.com", FirstName: "Original", LastName: "User", Password: "somepassword", // Password will be hashed by HashPassword
	}
	_ = originalUser.HashPassword() // Pre-hash for consistent test data if needed by CheckPassword later, though Update doesn't use it.


	t.Run("Success", func(t *testing.T) {
		updateParams := domainUser.UpdateUserParams{FirstName: "UpdatedFirst", LastName: "UpdatedLast"}
		// Reset user state for this test if necessary, or use a fresh one.
		// For this test, assume originalUser is the state before Update is called.
		// The GetByID mock should return this pre-update state.
		userBeforeUpdate := &domainUser.User{
			ID: originalUserID, Email: "original@example.com", FirstName: "Original", LastName: "User", Password: originalUser.Password, // Use already hashed password
		}
		mockRepo.On("GetByID", ctx, originalUserID).Return(userBeforeUpdate, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(u *domainUser.User) bool {
			return u.ID == originalUserID && u.FirstName == "UpdatedFirst" && u.LastName == "UpdatedLast"
		})).Return(nil).Once()

		updatedUser, err := userService.Update(ctx, originalUserID, updateParams)
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, "UpdatedFirst", updatedUser.FirstName)
		assert.Equal(t, "UpdatedLast", updatedUser.LastName)
		assert.Equal(t, userBeforeUpdate.Email, updatedUser.Email) // Email should not change in this case
		mockRepo.AssertExpectations(t)
	})

	t.Run("Email In Use", func(t *testing.T) {
		conflictingEmail := "taken@example.com"
		updateParams := domainUser.UpdateUserParams{Email: conflictingEmail}

		userForGetByID := &domainUser.User{ID: originalUserID, Email: "original@example.com", Password: "hashed"}
		conflictingUser := &domainUser.User{ID: uuid.New(), Email: conflictingEmail}

		mockRepo.On("GetByID", ctx, originalUserID).Return(userForGetByID, nil).Once()
		mockRepo.On("GetByEmail", ctx, conflictingEmail).Return(conflictingUser, nil).Once()

		_, err := userService.Update(ctx, originalUserID, updateParams)
		assert.Error(t, err)
		assert.EqualError(t, err, ErrEmailInUse.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Email In Use - GetByEmail returns gorm.ErrRecordNotFound for current user's email change to available", func(t *testing.T) {
		newEmail := "newavailable@example.com"
		updateParams := domainUser.UpdateUserParams{Email: newEmail, FirstName: "NewFirst"}

		userForGetByID := &domainUser.User{ID: originalUserID, Email: "original@example.com", Password: "hashed"}
		mockRepo.On("GetByID", ctx, originalUserID).Return(userForGetByID, nil).Once()
		mockRepo.On("GetByEmail", ctx, newEmail).Return(nil, gorm.ErrRecordNotFound).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(u *domainUser.User) bool {
			return u.ID == originalUserID && u.Email == newEmail && u.FirstName == "NewFirst"
		})).Return(nil).Once()

		updatedUser, err := userService.Update(ctx, originalUserID, updateParams)
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, newEmail, updatedUser.Email)
		assert.Equal(t, "NewFirst", updatedUser.FirstName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		nonExistentID := uuid.New()
		updateParams := domainUser.UpdateUserParams{FirstName: "Nobody"}
		mockRepo.On("GetByID", ctx, nonExistentID).Return(nil, nil).Once()

		_, err := userService.Update(ctx, nonExistentID, updateParams)
		assert.Error(t, err)
		assert.EqualError(t, err, ErrUserNotFound.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error on GetByID", func(t *testing.T) {
		updateParams := domainUser.UpdateUserParams{FirstName: "ErrorCase"}
		dbError := errors.New("db error on getbyid")
		mockRepo.On("GetByID", ctx, originalUserID).Return(nil, dbError).Once()

		_, err := userService.Update(ctx, originalUserID, updateParams)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user for update")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error on Update", func(t *testing.T) {
		updateParams := domainUser.UpdateUserParams{FirstName: "UpdateFail"}
		dbError := errors.New("db error on update")
		userForGetByID := &domainUser.User{ID: originalUserID, Email: "original@example.com", Password: "hashed"}
		mockRepo.On("GetByID", ctx, originalUserID).Return(userForGetByID, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*user.User")).Return(dbError).Once()

		_, err := userService.Update(ctx, originalUserID, updateParams)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user")
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdatePassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	currentPassword := "currentPassword123"
	newPassword := "newPassword456"

	testUser := &domainUser.User{ID: userID, Email: "user@example.com", Password: currentPassword}
	errHashing := testUser.HashPassword()
	assert.NoError(t, errHashing)

	t.Run("Success", func(t *testing.T) {
		// Ensure testUser has the correctly hashed currentPassword before this test
		userForGetByID := &domainUser.User{ID: userID, Email: "user@example.com", Password: testUser.Password}

		mockRepo.On("GetByID", ctx, userID).Return(userForGetByID, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(u *domainUser.User) bool {
			return u.ID == userID && bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(newPassword)) == nil
		})).Return(nil).Once()

		err := userService.UpdatePassword(ctx, userID, currentPassword, newPassword)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		nonExistentID := uuid.New()
		mockRepo.On("GetByID", ctx, nonExistentID).Return(nil, nil).Once()

		err := userService.UpdatePassword(ctx, nonExistentID, currentPassword, newPassword)
		assert.Error(t, err)
		assert.EqualError(t, err, ErrUserNotFound.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Incorrect Current Password", func(t *testing.T) {
		userForGetByID := &domainUser.User{ID: userID, Email: "user@example.com", Password: testUser.Password}
		mockRepo.On("GetByID", ctx, userID).Return(userForGetByID, nil).Once()

		err := userService.UpdatePassword(ctx, userID, "wrongCurrentPassword", newPassword)
		assert.Error(t, err)
		assert.EqualError(t, err, ErrIncorrectPassword.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error on GetByID", func(t *testing.T) {
		dbError := errors.New("db error on getbyid")
		mockRepo.On("GetByID", ctx, userID).Return(nil, dbError).Once()

		err := userService.UpdatePassword(ctx, userID, currentPassword, newPassword)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user for password update")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error on Update", func(t *testing.T) {
		dbError := errors.New("db error on update")
		userForGetByID := &domainUser.User{ID: userID, Email: "user@example.com", Password: testUser.Password}
		mockRepo.On("GetByID", ctx, userID).Return(userForGetByID, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*user.User")).Return(dbError).Once()

		err := userService.UpdatePassword(ctx, userID, currentPassword, newPassword)
		assert.Error(t, err)
		// This error message comes from the s.userRepo.Update(ctx, existingUser) call in user_service.go
		assert.Contains(t, err.Error(), "failed to update user")
		mockRepo.AssertExpectations(t)
	})
}
