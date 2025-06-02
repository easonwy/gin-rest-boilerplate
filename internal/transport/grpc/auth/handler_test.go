package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth" // Alias for domain auth types
	pb "github.com/yi-tech/go-user-service/api/proto/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockAuthService is a mock implementation of the auth.AuthService interface
type MockAuthService struct {
	mock.Mock
}

// Login mocks the Login method
func (m *MockAuthService) Login(ctx context.Context, email, password string) (*domainAuth.TokenPair, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainAuth.TokenPair), args.Error(1)
}

// RefreshToken mocks the RefreshToken method
func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*domainAuth.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainAuth.TokenPair), args.Error(1)
}

// Logout mocks the Logout method
func (m *MockAuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// ValidateToken mocks the ValidateToken method
func (m *MockAuthService) ValidateToken(ctx context.Context, accessToken string) (uuid.UUID, error) {
	args := m.Called(ctx, accessToken)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// Helper function to create mock domain TokenPair
func createMockDomainTokenPair() *domainAuth.TokenPair {
	return &domainAuth.TokenPair{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
	}
}

func TestNewHandler(t *testing.T) {
	mockService := new(MockAuthService)
	logger := zaptest.NewLogger(t)
	
	handler := NewHandler(mockService, logger)
	
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.authService)
	assert.Equal(t, logger, handler.logger)
}

func TestLogin(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	
	tests := []struct {
		name          string
		request       *pb.LoginRequest
		setupMock     func(*MockAuthService)
		expectedCode  codes.Code
		checkResponse func(*pb.TokenResponse)
	}{
		{
			name: "Success",
			request: &pb.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(mockService *MockAuthService) {
				mockTokenPair := createMockDomainTokenPair()
				mockService.On("Login", mock.Anything, "test@example.com", "password123").Return(mockTokenPair, nil)
			},
			expectedCode: codes.OK,
			checkResponse: func(response *pb.TokenResponse) {
				assert.Equal(t, "mock-access-token", response.AccessToken)
				assert.Equal(t, "mock-refresh-token", response.RefreshToken)
			},
		},
		{
			name: "Missing Email",
			request: &pb.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			setupMock: func(mockService *MockAuthService) {
				// No mock setup needed as validation should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Missing Password",
			request: &pb.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			setupMock: func(mockService *MockAuthService) {
				// No mock setup needed as validation should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid Credentials",
			request: &pb.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("Login", mock.Anything, "test@example.com", "wrongpassword").Return(nil, errors.New("invalid credentials"))
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Internal Error",
			request: &pb.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("Login", mock.Anything, "test@example.com", "password123").Return(nil, errors.New("database error"))
			},
			expectedCode: codes.Internal,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh handler and mock for each test to avoid interference
			mockService := new(MockAuthService)
			handler := NewHandler(mockService, logger)
			
			// Setup the mock expectations
			tt.setupMock(mockService)
			
			response, err := handler.Login(ctx, tt.request)
			
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

func TestRefreshToken(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	
	tests := []struct {
		name          string
		request       *pb.RefreshTokenRequest
		setupMock     func(*MockAuthService)
		expectedCode  codes.Code
		checkResponse func(*pb.TokenResponse)
	}{
		{
			name: "Success",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupMock: func(mockService *MockAuthService) {
				mockTokenPair := createMockDomainTokenPair()
				mockService.On("RefreshToken", mock.Anything, "valid-refresh-token").Return(mockTokenPair, nil)
			},
			expectedCode: codes.OK,
			checkResponse: func(response *pb.TokenResponse) {
				assert.Equal(t, "mock-access-token", response.AccessToken)
				assert.Equal(t, "mock-refresh-token", response.RefreshToken)
			},
		},
		{
			name: "Missing Refresh Token",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "",
			},
			setupMock: func(mockService *MockAuthService) {
				// No mock setup needed as validation should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid Token",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "invalid-token",
			},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("RefreshToken", mock.Anything, "invalid-token").Return(nil, errors.New("invalid token"))
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Session Not Found",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "expired-token",
			},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("RefreshToken", mock.Anything, "expired-token").Return(nil, errors.New("session not found"))
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Internal Error",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("RefreshToken", mock.Anything, "valid-refresh-token").Return(nil, errors.New("database error"))
			},
			expectedCode: codes.Internal,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh handler and mock for each test to avoid interference
			mockService := new(MockAuthService)
			handler := NewHandler(mockService, logger)
			
			// Setup the mock expectations
			tt.setupMock(mockService)
			
			response, err := handler.RefreshToken(ctx, tt.request)
			
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

func TestLogout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Create a context with user ID metadata
	userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000123")
	md := metadata.New(map[string]string{"user-id": userID.String()})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	
	tests := []struct {
		name         string
		request      *pb.LogoutRequest
		setupContext func() context.Context
		setupMock    func(*MockAuthService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			request: &pb.LogoutRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupContext: func() context.Context {
				return ctx
			},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("Logout", mock.Anything, userID).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Missing Refresh Token",
			request: &pb.LogoutRequest{
				RefreshToken: "",
			},
			setupContext: func() context.Context {
				return ctx
			},
			setupMock: func(mockService *MockAuthService) {
				// No mock setup needed as validation should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "No User ID in Context",
			request: &pb.LogoutRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupContext: func() context.Context {
				return context.Background() // Empty context without user ID
			},
			setupMock: func(mockService *MockAuthService) {
				// No mock setup needed as context validation should fail
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Invalid User ID Format",
			request: &pb.LogoutRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupContext: func() context.Context {
				invalidMD := metadata.New(map[string]string{"user-id": "invalid-id"})
				return metadata.NewIncomingContext(context.Background(), invalidMD)
			},
			setupMock: func(mockService *MockAuthService) {
				// No mock setup needed as user ID parsing should fail
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Session Not Found",
			request: &pb.LogoutRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupContext: func() context.Context {
				return ctx
			},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("Logout", mock.Anything, userID).Return(errors.New("session not found"))
			},
			expectedCode: codes.OK, // Still returns OK even if session not found
		},
		{
			name: "Internal Error",
			request: &pb.LogoutRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupContext: func() context.Context {
				return ctx
			},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("Logout", mock.Anything, userID).Return(errors.New("database error"))
			},
			expectedCode: codes.Internal,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh handler and mock for each test to avoid interference
			mockService := new(MockAuthService)
			handler := NewHandler(mockService, logger)
			
			// Setup the context and mock expectations
			testCtx := tt.setupContext()
			tt.setupMock(mockService)
			
			response, err := handler.Logout(testCtx, tt.request)
			
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

func TestValidateToken(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	
	tests := []struct {
		name          string
		request       *pb.ValidateTokenRequest
		setupMock     func(*MockAuthService)
		expectedCode  codes.Code
		checkResponse func(*pb.ValidateTokenResponse)
	}{
		{
			name: "Success",
			request: &pb.ValidateTokenRequest{
				AccessToken: "valid-access-token",
			},
			setupMock: func(mockService *MockAuthService) {
				userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000123")
				mockService.On("ValidateToken", mock.Anything, "valid-access-token").Return(userID, nil)
			},
			expectedCode: codes.OK,
			checkResponse: func(response *pb.ValidateTokenResponse) {
				assert.Equal(t, "00000000-0000-0000-0000-000000000123", response.UserId)
			},
		},
		{
			name: "Missing Access Token",
			request: &pb.ValidateTokenRequest{
				AccessToken: "",
			},
			setupMock: func(mockService *MockAuthService) {
				// No mock setup needed as validation should fail
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid Token",
			request: &pb.ValidateTokenRequest{
				AccessToken: "invalid-token",
			},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("ValidateToken", mock.Anything, "invalid-token").Return(uuid.Nil, errors.New("invalid token"))
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Internal Error",
			request: &pb.ValidateTokenRequest{
				AccessToken: "valid-access-token",
			},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("ValidateToken", mock.Anything, "valid-access-token").Return(uuid.Nil, errors.New("database error"))
			},
			expectedCode: codes.Internal,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh handler and mock for each test to avoid interference
			mockService := new(MockAuthService)
			handler := NewHandler(mockService, logger)
			
			// Setup the mock expectations
			tt.setupMock(mockService)
			
			response, err := handler.ValidateToken(ctx, tt.request)
			
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
