package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/yi-tech/go-user-service/api/proto/user"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
)

// MockUserService is a mock implementation of the domainUser.Service interface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, email, password, firstName, lastName string) (*domainUser.User, error) {
	args := m.Called(ctx, email, password, firstName, lastName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domainUser.User, error) {
	args := m.Called(ctx, id, firstName, lastName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserService) UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error {
	args := m.Called(ctx, id, currentPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func createMockUser() *domainUser.User {
	return &domainUser.User{
		ID:        uuid.New(), // Or a fixed test UUID: uuid.MustParse("your-test-uuid-here")
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestNewHandler(t *testing.T) {
	mockService := new(MockUserService)
	logger := zaptest.NewLogger(t)
	
	handler := NewHandler(mockService, logger)
	
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.userService)
	assert.Equal(t, logger, handler.logger)
}

func TestRegister(t *testing.T) {
	mockService := new(MockUserService)
	logger := zaptest.NewLogger(t)
	handler := NewHandler(mockService, logger)
	ctx := context.Background()
	
	tests := []struct {
		name          string
		request       *pb.RegisterRequest
		setupMock     func()
		expectedCode  codes.Code
		expectedUser  *pb.User
		checkResponse func(*pb.User)
	}{
		{
			name: "Success",
			request: &pb.RegisterRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "Test",
				LastName:  "User",
			},
			setupMock: func() {
				mockUser := createMockUser()
				mockService.On("Register", ctx, "test@example.com", "password123", "Test", "User").Return(mockUser, nil)
			},
			expectedCode: codes.OK,
			checkResponse: func(user *pb.User) {
				assert.NotEmpty(t, user.Id) // UUID format
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, "Test", user.FirstName)
			},
		},
		{
			name: "Missing Required Fields",
			request: &pb.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				// Missing FirstName
			},
			setupMock: func() {
				// No mock setup needed as validation should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "User Already Exists",
			request: &pb.RegisterRequest{
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "Existing",
				LastName:  "User",
			},
			setupMock: func() {
				mockService.On("Register", ctx, "existing@example.com", "password123", "Existing", "User").
					Return(nil, errors.New("user already exists"))
			},
			expectedCode: codes.AlreadyExists,
		},
		{
			name: "Internal Error",
			request: &pb.RegisterRequest{
				Email:     "error@example.com",
				Password:  "password123",
				FirstName: "Error",
				LastName:  "User",
			},
			setupMock: func() {
				mockService.On("Register", ctx, "error@example.com", "password123", "Error", "User").
					Return(nil, errors.New("database error"))
			},
			expectedCode: codes.Internal,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			
			response, err := handler.Register(ctx, tt.request)
			
			if tt.expectedCode != codes.OK {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				tt.checkResponse(response)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	
	validUUID := uuid.New()
	
	tests := []struct {
		name          string
		request       *pb.GetUserByIDRequest
		setupMock     func(*MockUserService)
		expectedCode  codes.Code
		checkResponse func(*pb.User)
	}{
		{
			name: "Success",
			request: &pb.GetUserByIDRequest{
				Id: validUUID.String(),
			},
			setupMock: func(mockService *MockUserService) {
				mockUser := createMockUser()
				mockService.On("GetByID", ctx, validUUID).Return(mockUser, nil)
			},
			expectedCode: codes.OK,
			checkResponse: func(user *pb.User) {
				assert.NotEmpty(t, user.Id) // UUID format
				assert.Equal(t, "test@example.com", user.Email)
			},
		},
		{
			name: "Invalid UUID",
			request: &pb.GetUserByIDRequest{
				Id: "invalid-uuid",
			},
			setupMock: func(mockService *MockUserService) {
				// No mock setup needed as UUID parsing should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "User Not Found",
			request: &pb.GetUserByIDRequest{
				Id: validUUID.String(),
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("GetByID", ctx, validUUID).Return(nil, errors.New("user not found"))
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Internal Error",
			request: &pb.GetUserByIDRequest{
				Id: validUUID.String(),
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("GetByID", ctx, validUUID).Return(nil, errors.New("database error"))
			},
			expectedCode: codes.Internal,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh handler and mock for each test to avoid interference
			mockService := new(MockUserService)
			handler := NewHandler(mockService, logger)
			
			// Setup the mock expectations
			tt.setupMock(mockService)
			
			response, err := handler.GetUserByID(ctx, tt.request)
			
			if tt.expectedCode != codes.OK {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				tt.checkResponse(response)
			}
			
			// Verify that all expected mock calls were made
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	mockService := new(MockUserService)
	logger := zaptest.NewLogger(t)
	handler := NewHandler(mockService, logger)
	ctx := context.Background()
	
	tests := []struct {
		name          string
		request       *pb.GetUserByEmailRequest
		setupMock     func()
		expectedCode  codes.Code
		checkResponse func(*pb.User)
	}{
		{
			name: "Success",
			request: &pb.GetUserByEmailRequest{
				Email: "test@example.com",
			},
			setupMock: func() {
				mockUser := createMockUser()
				mockService.On("GetByEmail", ctx, "test@example.com").Return(mockUser, nil)
			},
			expectedCode: codes.OK,
			checkResponse: func(user *pb.User) {
				assert.NotEmpty(t, user.Id) // UUID format
				assert.Equal(t, "test@example.com", user.Email)
			},
		},
		{
			name: "Empty Email",
			request: &pb.GetUserByEmailRequest{
				Email: "",
			},
			setupMock: func() {
				// No mock setup needed as validation should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "User Not Found",
			request: &pb.GetUserByEmailRequest{
				Email: "notfound@example.com",
			},
			setupMock: func() {
				mockService.On("GetByEmail", ctx, "notfound@example.com").Return(nil, errors.New("user not found"))
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Internal Error",
			request: &pb.GetUserByEmailRequest{
				Email: "error@example.com",
			},
			setupMock: func() {
				mockService.On("GetByEmail", ctx, "error@example.com").Return(nil, errors.New("database error"))
			},
			expectedCode: codes.Internal,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			
			response, err := handler.GetUserByEmail(ctx, tt.request)
			
			if tt.expectedCode != codes.OK {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				tt.checkResponse(response)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	
	validUUID := uuid.New()
	
	tests := []struct {
		name          string
		request       *pb.UpdateUserRequest
		setupMock     func(*MockUserService)
		expectedCode  codes.Code
		checkResponse func(*pb.User)
	}{
		{
			name: "Success",
			request: &pb.UpdateUserRequest{
				Id:        validUUID.String(),
				FirstName: "Updated",
				LastName:  "User",
			},
			setupMock: func(mockService *MockUserService) {
				updatedUser := createMockUser() // ID will be a new random UUID
				updatedUser.ID = validUUID // Ensure the mock returns the expected ID
				updatedUser.FirstName = "Updated"
				updatedUser.LastName = "User" // Assuming LastName is also part of the update or should match createMockUser
				mockService.On("UpdateUser", ctx, validUUID, "Updated", "User").Return(updatedUser, nil)
			},
			expectedCode: codes.OK,
			checkResponse: func(user *pb.User) {
				assert.Equal(t, validUUID.String(), user.Id)
				assert.Equal(t, "Updated", user.FirstName)
				assert.Equal(t, "User", user.LastName) // Add assertion for LastName
			},
		},
		{
			name: "Invalid UUID",
			request: &pb.UpdateUserRequest{
				Id:        "invalid-uuid",
				FirstName: "Updated",
				LastName:  "User",
			},
			setupMock: func(mockService *MockUserService) {
				// No mock setup needed as UUID parsing should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Missing First Name",
			request: &pb.UpdateUserRequest{
				Id:       validUUID.String(),
				LastName: "User",
			},
			setupMock: func(mockService *MockUserService) {
				// No mock setup needed as validation should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "User Not Found",
			request: &pb.UpdateUserRequest{
				Id:        validUUID.String(),
				FirstName: "Updated",
				LastName:  "User",
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("UpdateUser", ctx, validUUID, "Updated", "User").Return(nil, errors.New("user not found"))
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Internal Error",
			request: &pb.UpdateUserRequest{
				Id:        validUUID.String(),
				FirstName: "Updated",
				LastName:  "User",
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("UpdateUser", ctx, validUUID, "Updated", "User").Return(nil, errors.New("database error"))
			},
			expectedCode: codes.Internal,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh handler and mock for each test to avoid interference
			mockService := new(MockUserService)
			handler := NewHandler(mockService, logger)
			
			// Setup the mock expectations
			tt.setupMock(mockService)
			
			response, err := handler.UpdateUser(ctx, tt.request)
			
			if tt.expectedCode != codes.OK {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				tt.checkResponse(response)
			}
			
			// Verify that all expected mock calls were made
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdatePassword(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	
	validUUID := uuid.New()
	
	tests := []struct {
		name         string
		request      *pb.UpdatePasswordRequest
		setupMock    func(*MockUserService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			request: &pb.UpdatePasswordRequest{
				Id:              validUUID.String(),
				CurrentPassword: "oldpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("UpdatePassword", ctx, validUUID, "oldpassword", "newpassword").Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Invalid UUID",
			request: &pb.UpdatePasswordRequest{
				Id:              "invalid-uuid",
				CurrentPassword: "oldpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(mockService *MockUserService) {
				// No mock setup needed as UUID parsing should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Missing Passwords",
			request: &pb.UpdatePasswordRequest{
				Id:              validUUID.String(),
				CurrentPassword: "",
				NewPassword:     "",
			},
			setupMock: func(mockService *MockUserService) {
				// No mock setup needed as validation should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "User Not Found",
			request: &pb.UpdatePasswordRequest{
				Id:              validUUID.String(),
				CurrentPassword: "oldpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("UpdatePassword", ctx, validUUID, "oldpassword", "newpassword").Return(errors.New("user not found"))
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Internal Error",
			request: &pb.UpdatePasswordRequest{
				Id:              validUUID.String(),
				CurrentPassword: "oldpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("UpdatePassword", ctx, validUUID, "oldpassword", "newpassword").Return(errors.New("database error"))
			},
			expectedCode: codes.Internal,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh handler and mock for each test to avoid interference
			mockService := new(MockUserService)
			handler := NewHandler(mockService, logger)
			
			// Setup the mock expectations
			tt.setupMock(mockService)
			
			response, err := handler.UpdatePassword(ctx, tt.request)
			
			if tt.expectedCode != codes.OK {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.IsType(t, &emptypb.Empty{}, response)
			}
			
			// Verify that all expected mock calls were made
			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	
	validUUID := uuid.New()
	
	tests := []struct {
		name         string
		request      *pb.DeleteUserRequest
		setupMock    func(*MockUserService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			request: &pb.DeleteUserRequest{
				Id: validUUID.String(),
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("DeleteUser", ctx, validUUID).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Invalid UUID",
			request: &pb.DeleteUserRequest{
				Id: "invalid-uuid",
			},
			setupMock: func(mockService *MockUserService) {
				// No mock setup needed as UUID parsing should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "User Not Found",
			request: &pb.DeleteUserRequest{
				Id: validUUID.String(),
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("DeleteUser", ctx, validUUID).Return(errors.New("user not found"))
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Internal Error",
			request: &pb.DeleteUserRequest{
				Id: validUUID.String(),
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("DeleteUser", ctx, validUUID).Return(errors.New("database error"))
			},
			expectedCode: codes.Internal,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh handler and mock for each test to avoid interference
			mockService := new(MockUserService)
			handler := NewHandler(mockService, logger)
			
			// Setup the mock expectations
			tt.setupMock(mockService)
			
			response, err := handler.DeleteUser(ctx, tt.request)
			
			if tt.expectedCode != codes.OK {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.IsType(t, &emptypb.Empty{}, response)
			}
			
			// Verify that all expected mock calls were made
			mockService.AssertExpectations(t)
		})
	}
}

func TestToProtoUser(t *testing.T) {
	now := time.Now()
	testUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	domainUser := &domainUser.User{
		ID:        testUUID,
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	protoUser := toProtoUser(domainUser)

	assert.Equal(t, testUUID.String(), protoUser.Id)
	assert.Equal(t, "test@example.com", protoUser.Email)
	assert.Equal(t, "Test", protoUser.FirstName)
	assert.Equal(t, "User", protoUser.LastName)
	assert.Equal(t, now.Unix(), protoUser.CreatedAt.AsTime().Unix())
	assert.Equal(t, now.Unix(), protoUser.UpdatedAt.AsTime().Unix())
}
