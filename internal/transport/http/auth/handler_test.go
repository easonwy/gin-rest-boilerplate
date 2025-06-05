package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth" // Alias for domain auth types
	serviceAuth "github.com/yi-tech/go-user-service/internal/service/auth" // Import for sentinel errors
	"go.uber.org/zap/zaptest"
)

// MockAuthService is a mock of AuthService interface
type MockAuthService struct {
	mock.Mock
}

// Login mocks the Login method.
func (m *MockAuthService) Login(ctx context.Context, input domainAuth.LoginInput) (*domainAuth.TokenPair, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainAuth.TokenPair), args.Error(1)
}

// RefreshToken mocks the RefreshToken method.
func (m *MockAuthService) RefreshToken(ctx context.Context, token string) (*domainAuth.TokenPair, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainAuth.TokenPair), args.Error(1)
}

// Logout mocks the Logout method.
func (m *MockAuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// ValidateToken mocks the ValidateToken method.
// This method is part of the auth.AuthService interface but not directly used by this HTTP handler.
// We include it to fully implement the interface for the mock.
func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (uuid.UUID, error) {
	args := m.Called(ctx, token)
	if args.Error(1) != nil {
		return uuid.Nil, args.Error(1)
	}
	if args.Get(0) == nil {
		return uuid.Nil, args.Error(1)
	}
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// createMockTokenPair is a helper function to create a mock domainAuth.TokenPair for testing
func createMockTokenPair() *domainAuth.TokenPair {
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
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	mockTokenPair := createMockTokenPair() // Use new helper

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(mockService *MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			body: gin.H{"email": "test@example.com", "password": "password"},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("Login", mock.Anything, domainAuth.LoginInput{Email: "test@example.com", Password: "password"}).Return(mockTokenPair, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"code":200,"message":"Success","data":{"accessToken":"mock-access-token","refreshToken":"mock-refresh-token","expiresIn":3600}}`,
		},
		{
			name:           "Invalid Request Data - Bad JSON",
			body:           `{"email": "test@example.com", "password": "password"`, // Malformed JSON
			setupMock:      func(mockService *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"code":400,"message":"Invalid request data"}`,
		},
		{
			name: "Invalid Request Data - Missing Fields",
			body: gin.H{"email": "test@example.com"}, // Missing password
			setupMock: func(mockService *MockAuthService) {
				// No mock call expected as ShouldBindJSON should fail first
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"code":400,"message":"Invalid request data"}`,
		},
		{
			name: "Invalid Credentials",
			body: gin.H{"email": "wrong@example.com", "password": "wrong"},
			setupMock: func(mockService *MockAuthService) {
				// Use the actual sentinel error from the service/auth package
				mockService.On("Login", mock.Anything, domainAuth.LoginInput{Email: "wrong@example.com", Password: "wrong"}).Return(nil, serviceAuth.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			// The message should now match ErrInvalidCredentials.Error()
			expectedBody:   `{"code":401,"message":"invalid credentials"}`,
		},
		{
			name: "Internal ServerError",
			body: gin.H{"email": "error@example.com", "password": "password"},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("Login", mock.Anything, domainAuth.LoginInput{Email: "error@example.com", Password: "password"}).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"code":500,"message":"Something went wrong. Please try again later."}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tc.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			rr := httptest.NewRecorder()
			ctx, router := gin.CreateTestContext(rr)
			router.POST("/login", handler.Login)

			var reqBodyReader io.Reader
			if tc.body != nil {
				if strBody, ok := tc.body.(string); ok {
					reqBodyReader = strings.NewReader(strBody)
				} else {
					jsonBody, _ := json.Marshal(tc.body)
					reqBodyReader = bytes.NewBuffer(jsonBody)
				}
			}

			req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/login", reqBodyReader)
			if tc.body != nil && (tc.name != "Invalid Request Data - Bad JSON") {
				req.Header.Set("Content-Type", "application/json")
			}

			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	mockTokenPair := createMockTokenPair() // Use new helper

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(mockService *MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			body: gin.H{"refreshToken": "valid-refresh-token"},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("RefreshToken", mock.AnythingOfType("*gin.Context"), "valid-refresh-token").Return(mockTokenPair, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"code":200,"message":"Success","data":{"accessToken":"mock-access-token","refreshToken":"mock-refresh-token","expiresIn":3600}}`,
		},
		{
			name:           "Invalid Request Data - Bad JSON",
			body:           `{"refresh_token": "valid-refresh-token"`, // Malformed JSON
			setupMock:      func(mockService *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"code":400,"message":"Invalid request data"}`,
		},
		{
			name: "Invalid Request Data - Missing Refresh Token",
			body: gin.H{}, // Missing refresh_token
			setupMock: func(mockService *MockAuthService) {
				// No mock call expected as ShouldBindJSON should fail first
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"code":400,"message":"Invalid request data"}`,
		},
		{
			name: "Invalid or Expired Token",
			body: gin.H{"refreshToken": "invalid-token"},
			setupMock: func(mockService *MockAuthService) {
				// Use the actual sentinel error
				mockService.On("RefreshToken", mock.AnythingOfType("*gin.Context"), "invalid-token").Return(nil, serviceAuth.ErrInvalidOrExpiredToken)
			},
			expectedStatus: http.StatusUnauthorized,
			// The message should now match ErrInvalidOrExpiredToken.Error()
			expectedBody:   `{"code":401,"message":"invalid or expired refresh token"}`,
		},
		{
			name: "Internal Server Error on Refresh",
			body: gin.H{"refreshToken": "error-token"},
			setupMock: func(mockService *MockAuthService) {
				mockService.On("RefreshToken", mock.AnythingOfType("*gin.Context"), "error-token").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"code":500,"message":"Something went wrong. Please try again later."}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tc.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			rr := httptest.NewRecorder()
			ctx, router := gin.CreateTestContext(rr)
			router.POST("/refresh", handler.RefreshToken) // Endpoint for refresh token

			var reqBodyReader io.Reader
			if tc.body != nil {
				if strBody, ok := tc.body.(string); ok {
					reqBodyReader = strings.NewReader(strBody)
				} else {
					jsonBody, _ := json.Marshal(tc.body)
					reqBodyReader = bytes.NewBuffer(jsonBody)
				}
			}

			req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/refresh", reqBodyReader)
			if tc.body != nil && (tc.name != "Invalid Request Data - Bad JSON") {
				req.Header.Set("Content-Type", "application/json")
			}

			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		setupContext   func(c *gin.Context) // Changed to modify the handler's actual context
		setupMock      func(mockService *MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			setupContext: func(c *gin.Context) {
				userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000123")
				c.Set("userID", userID) // Corrected context key
			},
			setupMock: func(mockService *MockAuthService) {
				userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000123")
				mockService.On("Logout", mock.Anything, userID).Return(nil) // mock.Anything matches context.Context
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"code":200,"message":"Success","data":{"message":"Logged out successfully"}}`,
		},
		{
			name:           "Authentication Required - No User ID in Context",
			setupContext:   nil, // No context setup needed, or func(c *gin.Context) {}
			setupMock:      func(mockService *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"code":401,"message":"Authentication required"}`,
		},
		{
			name: "Internal Server Error - Invalid User ID Type in Context",
			setupContext: func(c *gin.Context) {
				c.Set("userID", "not-a-uuid") // Corrected context key, set user_id as string
			},
			setupMock:      func(mockService *MockAuthService) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"code":500,"message":"Something went wrong. Please try again later."}`,
		},
		{
			name: "Internal Server Error - Logout Fails",
			setupContext: func(c *gin.Context) {
				userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000123")
				c.Set("userID", userID) // Corrected context key
			},
			setupMock: func(mockService *MockAuthService) {
				userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000123")
				// Mocking a generic error for this case, as the handler should catch it and return 500
				mockService.On("Logout", mock.Anything, userID).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"code":500,"message":"Something went wrong. Please try again later."}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tc.setupMock(mockService)

			handler := NewHandler(mockService, logger)
			rr := httptest.NewRecorder()
			// We only need the router from CreateTestContext. The returned context is not used directly for setting handler keys.
			_, router := gin.CreateTestContext(rr)

			// Define the endpoint with a wrapper that sets up the context for the handler
			router.POST("/logout", func(c *gin.Context) {
				if tc.setupContext != nil {
					tc.setupContext(c) // Apply context modifications to the actual handler's context
				}
				handler.Logout(c) // Call the actual handler
			})

			// Create a plain request. The context of this request (req.Context()) will be available
			// to the service via c.Request.Context(), and is what mock.AnythingOfType("context.Context") matches.
			req, _ := http.NewRequest(http.MethodPost, "/logout", nil)

			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
