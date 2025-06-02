package user

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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// stringPtr is a helper function to get a pointer to a string. Useful for DTOs with optional string fields.
func stringPtr(s string) *string {
	return &s
}

// MockUserService is a mock type for the UserService interface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, email, password, firstName, lastName string) (*domainUser.User, error) {
	args := m.Called(mock.Anything, email, password, firstName, lastName) // Use mock.Anything for ctx
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error) {
	args := m.Called(mock.Anything, id) // Use mock.Anything for ctx
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	args := m.Called(mock.Anything, email) // Use mock.Anything for ctx
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domainUser.User, error) {
	args := m.Called(mock.Anything, id, firstName, lastName) // Use mock.Anything for ctx
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserService) UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error {
	args := m.Called(mock.Anything, id, currentPassword, newPassword) // Use mock.Anything for ctx
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(mock.Anything, id) // Use mock.Anything for ctx
	return args.Error(0)
}

// createMockDomainUser creates a mock domainUser.User object for testing.
func createMockDomainUser(id uuid.UUID, email, firstName, lastName string) *domainUser.User {
	now := time.Now()
	return &domainUser.User{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  "hashedpassword", // Not directly used in response, but good for struct completeness
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestNewUserHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := new(MockUserService)
	handler := NewHandler(mockService, logger) // NewHandler is from the package user, not user_test

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.userService)
	assert.Equal(t, logger, handler.logger)
}

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	mockUserID := uuid.New()
	mockUser := createMockDomainUser(mockUserID, "test@example.com", "Test", "User")
	// Expected response based on toUserResponse and current domainUser.User structure
	expectedUserResponse := UserResponse{
		ID:        mockUser.ID.String(),
		Email:     mockUser.Email,
		FirstName: mockUser.FirstName,
		LastName:  mockUser.LastName,
		CreatedAt: mockUser.CreatedAt, // Direct time.Time comparison
		UpdatedAt: mockUser.UpdatedAt, // Direct time.Time comparison
	}
	expectedUserJSON, _ := json.Marshal(expectedUserResponse)
	// For successful registration, we'll unmarshal the response and compare structs
	// to handle time.Time precision issues if comparing JSON strings directly.

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(mockService *MockUserService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			body: gin.H{
				"email":      "test@example.com",
				"password":   "password123",
				"first_name": "Test",
				"last_name":  "User",
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("Register", mock.Anything, "test@example.com", "password123", "Test", "User").Return(mockUser, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   string(expectedUserJSON),
		},
		{
			name:           "Invalid Request Data - Bad JSON",
			body:           `{"email": "test@example.com", "password": "password123"`, // Malformed JSON
			setupMock:      func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request data"}`,
		},
		{
			name: "Invalid Request Data - Missing Fields",
			body: gin.H{"email": "test@example.com"}, // Missing password, first_name, last_name
			setupMock: func(mockService *MockUserService) {
				// No mock call expected as ShouldBindJSON should fail first
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request data"}`,
		},
		{
			name: "User Already Exists",
			body: gin.H{
				"email":      "existing@example.com",
				"password":   "password123",
				"first_name": "Existing",
				"last_name":  "User",
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("Register", mock.Anything, "existing@example.com", "password123", "Existing", "User").Return(nil, errors.New("user already exists"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"Email already exists"}`,
		},
		{
			name: "Internal Server Error",
			body: gin.H{
				"email":      "error@example.com",
				"password":   "password123",
				"first_name": "Error",
				"last_name":  "User",
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("Register", mock.Anything, "error@example.com", "password123", "Error", "User").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to register user"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tc.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			rr := httptest.NewRecorder()
			ctx, router := gin.CreateTestContext(rr)
			router.POST("/register", handler.Register)

			var reqBodyReader *bytes.Buffer
			if tc.body != nil {
				if strBody, ok := tc.body.(string); ok {
					reqBodyReader = bytes.NewBufferString(strBody)
				} else {
					jsonBody, _ := json.Marshal(tc.body)
					reqBodyReader = bytes.NewBuffer(jsonBody)
				}
			}

			req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/register", reqBodyReader)
			if tc.body != nil && reqBodyReader != nil && (tc.name != "Invalid Request Data - Bad JSON") {
				req.Header.Set("Content-Type", "application/json")
			}

			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdatePassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	mockUserUUID := uuid.New()

	tests := []struct {
		name           string
		userIDParam    string
		requestBody    interface{}
		setupMock      func(mockService *MockUserService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			userIDParam: mockUserUUID.String(),
			requestBody: UpdatePasswordRequest{
				CurrentPassword: "oldPassword123",
				NewPassword:     "newPassword456",
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("UpdatePassword", mock.Anything, mockUserUUID, "oldPassword123", "newPassword456").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Password updated successfully"}`,
		},
		{
			name:        "Invalid User ID Format",
			userIDParam: "not-a-uuid",
			requestBody: UpdatePasswordRequest{CurrentPassword: "old", NewPassword: "newPasswordValid"},
			setupMock:   func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid user ID format"}`,
		},
		{
			name:        "Invalid Request Data - Malformed JSON",
			userIDParam: mockUserUUID.String(),
			requestBody: "not json",
			setupMock:   func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request data"}`,
		},
		{
			name:        "Invalid Request Data - Missing CurrentPassword",
			userIDParam: mockUserUUID.String(),
			requestBody: UpdatePasswordRequest{NewPassword: "newPasswordValid"},
			setupMock:   func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request data"}`,
		},
		{
			name:        "Invalid Request Data - Missing NewPassword",
			userIDParam: mockUserUUID.String(),
			requestBody: UpdatePasswordRequest{CurrentPassword: "oldPassword123"},
			setupMock:   func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request data"}`,
		},
		{
			name:        "Invalid Request Data - NewPassword too short",
			userIDParam: mockUserUUID.String(),
			requestBody: UpdatePasswordRequest{CurrentPassword: "oldPassword123", NewPassword: "short"},
			setupMock:   func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request data"}`,
		},
		{
			name:        "User Not Found",
			userIDParam: uuid.New().String(), // Different valid UUID
			requestBody: UpdatePasswordRequest{CurrentPassword: "old", NewPassword: "newPasswordValid"},
			setupMock: func(mockService *MockUserService) {
				mockService.On("UpdatePassword", mock.Anything, mock.AnythingOfType("uuid.UUID"), "old", "newPasswordValid").Return(errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"User not found"}`,
		},
		{
			name:        "Incorrect Current Password", // Assuming service layer returns a specific error or a generic one handled by handler
			userIDParam: mockUserUUID.String(),
			requestBody: UpdatePasswordRequest{CurrentPassword: "wrongOldPassword", NewPassword: "newPasswordValid"},
			setupMock: func(mockService *MockUserService) {
				// For this example, let's assume the service returns a generic error that the handler maps to InternalServerError
				// A more specific error like "invalid credentials" could be mapped to http.StatusUnauthorized or http.StatusBadRequest
				mockService.On("UpdatePassword", mock.Anything, mockUserUUID, "wrongOldPassword", "newPasswordValid").Return(errors.New("invalid current password"))
			},
			expectedStatus: http.StatusInternalServerError, // Or StatusBadRequest/StatusUnauthorized if handler maps it differently
			expectedBody:   `{"error":"Failed to update password"}`,
		},
		{
			name:        "Internal Server Error",
			userIDParam: mockUserUUID.String(),
			requestBody: UpdatePasswordRequest{CurrentPassword: "old", NewPassword: "newPasswordValid"},
			setupMock: func(mockService *MockUserService) {
				mockService.On("UpdatePassword", mock.Anything, mockUserUUID, "old", "newPasswordValid").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to update password"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tc.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			rr := httptest.NewRecorder()
			_, router := gin.CreateTestContext(rr)
			router.PUT("/users/:id/password", handler.UpdatePassword) // Assuming this path

			var reqBodyReader io.Reader
			if tc.requestBody != nil {
				if strBody, ok := tc.requestBody.(string); ok {
					reqBodyReader = strings.NewReader(strBody)
				} else {
					bodyBytes, _ := json.Marshal(tc.requestBody)
					reqBodyReader = bytes.NewReader(bodyBytes)
				}
			}

			req, _ := http.NewRequest(http.MethodPut, "/users/"+tc.userIDParam+"/password", reqBodyReader)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	mockUserUUID := uuid.New()

	tests := []struct {
		name           string
		userIDParam    string
		setupMock      func(mockService *MockUserService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			userIDParam: mockUserUUID.String(),
			setupMock: func(mockService *MockUserService) {
				mockService.On("DeleteUser", mock.Anything, mockUserUUID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"User deleted successfully"}`,
		},
		{
			name:        "Invalid User ID Format",
			userIDParam: "not-a-uuid",
			setupMock:   func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid user ID format"}`,
		},
		{
			name:        "User Not Found",
			userIDParam: uuid.New().String(), // Different valid UUID
			setupMock: func(mockService *MockUserService) {
				mockService.On("DeleteUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"User not found"}`,
		},
		{
			name:        "Internal Server Error",
			userIDParam: mockUserUUID.String(),
			setupMock: func(mockService *MockUserService) {
				mockService.On("DeleteUser", mock.Anything, mockUserUUID).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to delete user"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tc.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			rr := httptest.NewRecorder()
			_, router := gin.CreateTestContext(rr)
			router.DELETE("/users/:id", handler.DeleteUser) // Assuming this path

			req, _ := http.NewRequest(http.MethodDelete, "/users/"+tc.userIDParam, nil)
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	// mockUserUUID is the UUID used in context and service calls
	mockUserUUID := uuid.New()

	// Base user for tests
	baseDomainUser := createMockDomainUser(mockUserUUID, "test@example.com", "OriginalFirst", "OriginalLast")

	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		requestBody    string
		setupMock      func(mockService *MockUserService, userID uuid.UUID, reqBody UpdateRequest)
		expectedStatus int
		expectedBody   map[string]interface{}
		ignoreBody     bool
	}{
		{
			name: "Success",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", mockUserUUID.String())
			},
			requestBody: `{"first_name": "NewFirst", "last_name": "NewLast"}`,
			setupMock: func(mockService *MockUserService, userID uuid.UUID, reqBody UpdateRequest) {
				updatedTime := time.Now() // Simulate time of update
				returnedUserAfterUpdate := &domainUser.User{
					ID:        userID,
					Email:     baseDomainUser.Email, // Email is not changed in this operation
					FirstName: reqBody.FirstName,    // Updated FirstName
					LastName:  reqBody.LastName,     // Updated LastName
					Password:  baseDomainUser.Password, // Password not changed
					CreatedAt: baseDomainUser.CreatedAt, // CreatedAt not changed
					UpdatedAt: updatedTime,          // Service updates this
				}
				mockService.On("UpdateUser", mock.Anything, userID, reqBody.FirstName, reqBody.LastName).
					Return(returnedUserAfterUpdate, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         mockUserUUID.String(),
				"email":      baseDomainUser.Email,
				"first_name": "NewFirst", // Matches requestBody in the success test case
				"last_name":  "NewLast",  // Matches requestBody in the success test case
				"created_at": baseDomainUser.CreatedAt.Format(time.RFC3339),
				"updated_at": time.Now().Format(time.RFC3339),
			},
			ignoreBody: false, // We will assert the full body now
		},
		{
			name: "Unauthorized - no user_id in context",
			setupContext: func(c *gin.Context) {
				// No user_id set
			},
			requestBody:    `{"first_name": "NewFirst", "last_name": "NewLast"}`,
			setupMock:      func(mockService *MockUserService, userID uuid.UUID, reqBody UpdateRequest) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]interface{}{"error": "Unauthorized"},
		},
		{
			name: "Internal Server Error - invalid user_id string format in context",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", "not-a-uuid")
			},
			requestBody:    `{"first_name": "NewFirst", "last_name": "NewLast"}`,
			setupMock:      func(mockService *MockUserService, userID uuid.UUID, reqBody UpdateRequest) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "Internal server error"},
		},
		{
			name: "Internal Server Error - invalid user_id type in context",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 123)
			},
			requestBody:    `{"first_name": "NewFirst", "last_name": "NewLast"}`,
			setupMock:      func(mockService *MockUserService, userID uuid.UUID, reqBody UpdateRequest) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "Internal server error"},
		},
		{
			name: "Invalid Request Data - Malformed JSON",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", mockUserUUID.String()) // mockUserUUID needs to be defined in TestUpdateProfile scope
			},
			requestBody:    `{"first_name": "NewFirst", "last_name": "NewLast"`, // Malformed
			setupMock:      func(mockService *MockUserService, userID uuid.UUID, reqBody UpdateRequest) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "Invalid request data"},
		},
		{
			name: "Invalid Request Data - Missing FirstName",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", mockUserUUID.String()) // mockUserUUID needs to be defined
			},
			requestBody:    `{"last_name": "NewLast"}`,
			setupMock:      func(mockService *MockUserService, userID uuid.UUID, reqBody UpdateRequest) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "Invalid request data"},
		},
		{
			name: "User Not Found",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", mockUserUUID.String()) // mockUserUUID needs to be defined
			},
			requestBody: `{"first_name": "NewFirst", "last_name": "NewLast"}`,
			setupMock: func(mockService *MockUserService, userID uuid.UUID, reqBody UpdateRequest) {
				mockService.On("UpdateUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), reqBody.FirstName, reqBody.LastName).Return(nil, errors.New("user not found")).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "User not found"},
		},
		{
			name: "Internal Server Error - service error",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", mockUserUUID.String()) // mockUserUUID needs to be defined
			},
			requestBody: `{"first_name": "NewFirst", "last_name": "NewLast"}`,
			setupMock: func(mockService *MockUserService, userID uuid.UUID, reqBody UpdateRequest) {
				mockService.On("UpdateUser", mock.Anything, userID, reqBody.FirstName, reqBody.LastName).Return(nil, errors.New("database error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "Failed to update user profile"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockUserService)

			handler := NewHandler(mockService, logger)

			var reqBody UpdateRequest
			if tc.requestBody != "" {
				json.Unmarshal([]byte(tc.requestBody), &reqBody) // For setupMock
			}
			tc.setupMock(mockService, mockUserUUID, reqBody)

			rr := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rr)

			c.Request, _ = http.NewRequest(http.MethodPut, "/profile", strings.NewReader(tc.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			tc.setupContext(c)

			handler.UpdateProfile(c)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if !tc.ignoreBody {
				var responseBody map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBody, responseBody)
			} else {
				// If ignoreBody is true, but expectedBody is not nil, it implies a scenario
				// where we might want to do partial checks or format-specific checks not handled by direct equality.
				// For now, if ignoreBody is true, we don't assert the body content here.
				// Specific tests needing this can add custom assertions within their tc.setupMock or similar.
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	mockUserUUID_GetProfile := uuid.New()

	// Use helper to create a consistent domainUser.User for tests
	baseUserForGetProfile := createMockDomainUser(mockUserUUID_GetProfile, "profileget@example.com", "ProfileFirst", "ProfileLast")

	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		setupMock      func(mockService *MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
		ignoreBody     bool
	}{
		{
			name: "Success",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", mockUserUUID_GetProfile.String())
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("GetByID", mock.Anything, mockUserUUID_GetProfile).Return(baseUserForGetProfile, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         mockUserUUID_GetProfile.String(),
				"email":      baseUserForGetProfile.Email,
				"first_name": baseUserForGetProfile.FirstName,
				"last_name":  baseUserForGetProfile.LastName,
				"created_at": baseUserForGetProfile.CreatedAt.Format(time.RFC3339),
				"updated_at": baseUserForGetProfile.UpdatedAt.Format(time.RFC3339),
			},
			ignoreBody: false,
		},
		{
			name:           "Unauthorized - no user_id in context",
			setupContext:   func(c *gin.Context) {},
			setupMock:      func(mockService *MockUserService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]interface{}{"error": "Unauthorized"},
		},
		{
			name: "Internal Server Error - invalid user_id string format in context",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", "not-a-uuid")
			},
			setupMock:      func(mockService *MockUserService) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "Internal server error"},
		},
		{
			name: "Internal Server Error - invalid user_id type in context",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 123) // int type
			},
			setupMock:      func(mockService *MockUserService) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "Internal server error"},
		},
		{
			name: "User Not Found",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", mockUserUUID_GetProfile.String())
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("GetByID", mock.Anything, mockUserUUID_GetProfile).Return(nil, errors.New("user not found")).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "User not found"},
		},
		{
			name: "Internal Server Error - service error",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", mockUserUUID_GetProfile.String())
			},
			setupMock: func(mockService *MockUserService) {
				mockService.On("GetByID", mock.Anything, mockUserUUID_GetProfile).Return(nil, errors.New("database error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "Failed to get user profile"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tc.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			rr := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rr)
			c.Request, _ = http.NewRequest(http.MethodGet, "/profile", nil)
			tc.setupContext(c)

			handler.GetProfile(c)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if !tc.ignoreBody {
				var responseBody map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBody, responseBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	mockUserUUID := uuid.New()

	baseUser := createMockDomainUser(mockUserUUID, "original@example.com", "OriginalFirst", "OriginalLast")

	updatedFirstName := "UpdatedFirst"
	updatedLastName := "UpdatedLast"

	// This is the user object the service is expected to return after a successful update.
	successUserForMockReturn := &domainUser.User{
		ID:        mockUserUUID,
		Email:     baseUser.Email, // Email is not updated in this scenario
		FirstName: updatedFirstName,
		LastName:  updatedLastName,
		Password:  baseUser.Password, // Password not changed
		CreatedAt: baseUser.CreatedAt, // CreatedAt does not change
		UpdatedAt: time.Now(),       // UpdatedAt is set by the service
	}

	// Expected JSON structure for the success case response body
	expectedSuccessBodyMap := map[string]interface{}{
		"id":         mockUserUUID.String(),
		"email":      baseUser.Email,
		"first_name": updatedFirstName,
		"last_name":  updatedLastName,
		"created_at": baseUser.CreatedAt.Format(time.RFC3339),
		// UpdatedAt will be checked for presence and non-emptiness due to precision issues
	}

	tests := []struct {
		name           string
		userIDParam    string
		requestBody    interface{}
		setupMock      func(mockService *MockUserService)
		expectedStatus int
		expectedBody   string
		ignoreBody     bool // For cases where UpdatedAt makes exact match hard
	}{
		{
			name:        "Success",
			userIDParam: mockUserUUID.String(),
			requestBody: UserUpdateRequest{ // Use the DTO from user_dtos.go
				FirstName: &updatedFirstName,
				LastName:  &updatedLastName,
			},
			setupMock: func(mockService *MockUserService) {
				// Adjust the returned user's UpdatedAt to be predictable for the mock, or use mock.Anything for it if not strictly compared.
				// For simplicity, we'll use the successUserForMockReturn which has a fresh time.Now().
				mockService.On("UpdateUser", mock.Anything, mockUserUUID, updatedFirstName, updatedLastName).Return(successUserForMockReturn, nil).Once()
			},
			expectedStatus: http.StatusOK,
			// expectedBody will be compared as a map, see assertion logic below
			ignoreBody:     true, // Special handling for UpdatedAt
		},
		{
			name:        "Invalid User ID Format",
			userIDParam: "not-a-uuid",
			requestBody: UserUpdateRequest{FirstName: stringPtr("Test"), LastName: stringPtr("User")},
			setupMock:   func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid user ID format"}`,
		},
		{
			name:           "Invalid Request Data - Malformed JSON",
			userIDParam:    mockUserUUID.String(),
			requestBody:    "not json",
			setupMock:      func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request data"}`,
		},
		{
			name:        "Invalid Request Data - Missing FirstName",
			userIDParam: mockUserUUID.String(),
			requestBody: UserUpdateRequest{LastName: stringPtr("User")}, // Missing FirstName
			setupMock:   func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request data"}`,
		},
		// Note: UpdateUser in handler.go does not validate for missing LastName, so a test for it might pass if not required by service.
		// Assuming UpdateRequest has validation tags for both FirstName and LastName if they are required.
		// If LastName is optional, this test case for missing LastName might not be relevant or should expect success.
		{
			name:        "User Not Found",
			userIDParam: uuid.New().String(), // Different valid UUID
			requestBody: UserUpdateRequest{FirstName: stringPtr("Test"), LastName: stringPtr("User")},
			setupMock: func(mockService *MockUserService) {
				mockService.On("UpdateUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Test", "User").Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"User not found"}`,
		},
		{
			name:        "Internal Server Error",
			userIDParam: uuid.New().String(),
			requestBody: UserUpdateRequest{FirstName: stringPtr("Test"), LastName: stringPtr("User")},
			setupMock: func(mockService *MockUserService) {
				mockService.On("UpdateUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Test", "User").Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tc.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			rr := httptest.NewRecorder()
			_, router := gin.CreateTestContext(rr)
			router.PUT("/users/:id", handler.UpdateUser) // Define route for param handling

			var bodyReader io.Reader
			if tc.requestBody != nil {
				if strBody, ok := tc.requestBody.(string); ok {
					bodyReader = strings.NewReader(strBody)
				} else {
					bodyBytes, err := json.Marshal(tc.requestBody)
					assert.NoError(t, err)
					bodyReader = bytes.NewBuffer(bodyBytes)
				}
			}

			req, err := http.NewRequest(http.MethodPut, "/users/"+tc.userIDParam, bodyReader)
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(rr, req) // Use router to serve the request for param handling

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			} else if tc.ignoreBody && tc.name == "Success" { // Special handling for success case with dynamic UpdatedAt
				var responseBodyMap map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &responseBodyMap)
				assert.NoError(t, err, "Failed to unmarshal actual response body for success case")

				for k, expectedValue := range expectedSuccessBodyMap { // Compare against the predefined map
					assert.Equal(t, expectedValue, responseBodyMap[k], "Field mismatch for key: %s", k)
				}
				// Check 'updated_at' separately for presence and format if needed (already formatted in toUserResponse)
				assert.Contains(t, responseBodyMap, "updated_at", "Response should contain 'updated_at' field")
				assert.NotEmpty(t, responseBodyMap["updated_at"], "'updated_at' field should not be empty")
				// Optionally, parse and check if it's a valid RFC3339 timestamp
				_, err = time.Parse(time.RFC3339, responseBodyMap["updated_at"].(string))
				assert.NoError(t, err, "updated_at field is not a valid RFC3339 timestamp")
			} // This closes the 'else if'
			// The 'if tc.expectedBody != ""' from line 830 was never closed, but the 'else if' implies it should be part of that structure.
			// The logic should be: if expectedBody is set, assert.JSONEq. ELSE (if ignoreBody and success), do map comparison.
			// The original structure was: if expectedBody { if expectedBody {} else if ignoreBody {}} which is wrong.
			// Corrected structure: if expectedBody AND NOT (ignoreBody AND success) { assert.JSONEq } else if ignoreBody AND success { map compare }
			// Simpler: if tc.name == "Success" && tc.ignoreBody { map compare } else if tc.expectedBody != "" { assert.JSONEq }

			// Let's re-evaluate the logic based on the intent:
			// 1. If it's the "Success" case AND ignoreBody is true, do the special map comparison for UpdatedAt.
			// 2. ELSE (if it's not the success/ignoreBody case), AND if expectedBody is not empty, do a normal JSONEq.

			if tc.name == "Success" && tc.ignoreBody {
				var responseBodyMap map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &responseBodyMap)
				assert.NoError(t, err, "Failed to unmarshal actual response body for success case")

				for k, expectedValue := range expectedSuccessBodyMap { // Compare against the predefined map
					assert.Equal(t, expectedValue, responseBodyMap[k], "Field mismatch for key: %s", k)
				}
				assert.Contains(t, responseBodyMap, "updated_at", "Response should contain 'updated_at' field")
				assert.NotEmpty(t, responseBodyMap["updated_at"], "'updated_at' field should not be empty")
				_, err = time.Parse(time.RFC3339, responseBodyMap["updated_at"].(string))
				assert.NoError(t, err, "updated_at field is not a valid RFC3339 timestamp")
			} else if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	mockUserUUID := uuid.New()
	mockUserEmail := "test@example.com"
	mockUser := createMockDomainUser(mockUserUUID, mockUserEmail, "Test", "User")

	expectedUserResponse := UserResponse{
		ID:        mockUser.ID.String(),
		Email:     mockUser.Email,
		FirstName: mockUser.FirstName,
		LastName:  mockUser.LastName,
		CreatedAt: mockUser.CreatedAt,
		UpdatedAt: mockUser.UpdatedAt,
	}
	expectedUserJSON, _ := json.Marshal(expectedUserResponse)

	tests := []struct {
		name           string
		emailQuery     string
		setupMock      func(mockService *MockUserService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "Success",
			emailQuery: "?email=" + mockUserEmail,
			setupMock: func(mockService *MockUserService) {
				mockService.On("GetByEmail", mock.Anything, mockUserEmail).Return(mockUser, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   string(expectedUserJSON),
		},
		{
			name:       "Missing Email Parameter",
			emailQuery: "", // No email query param
			setupMock:   func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Email parameter is required"}`,
		},
		{
			name:       "User Not Found",
			emailQuery: "?email=notfound@example.com",
			setupMock: func(mockService *MockUserService) {
				mockService.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"User not found"}`,
		},
		{
			name:       "Internal Server Error",
			emailQuery: "?email=error@example.com",
			setupMock: func(mockService *MockUserService) {
				mockService.On("GetByEmail", mock.Anything, "error@example.com").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get user"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tc.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			rr := httptest.NewRecorder()
			_, router := gin.CreateTestContext(rr)
			// The actual handler path might be different, e.g. /users/email or just /email under a group
			// Assuming /users/email for now based on typical patterns
			router.GET("/users/email", handler.GetUserByEmail) 

			req, _ := http.NewRequest(http.MethodGet, "/users/email"+tc.emailQuery, nil)
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
func TestGetUserByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	mockUserUUID := uuid.New()
	mockUser := createMockDomainUser(mockUserUUID, "test@example.com", "Test", "User")

	// Expected response based on toUserResponse and current domainUser.User structure
	expectedUserResponse := UserResponse{
		ID:        mockUser.ID.String(),
		Email:     mockUser.Email,
		FirstName: mockUser.FirstName,
		LastName:  mockUser.LastName,
		CreatedAt: mockUser.CreatedAt,
		UpdatedAt: mockUser.UpdatedAt,
	}
	expectedUserJSON, _ := json.Marshal(expectedUserResponse)

	tests := []struct {
		name           string
		userIDParam    string
		setupMock      func(mockService *MockUserService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			userIDParam: mockUserUUID.String(),
			setupMock: func(mockService *MockUserService) {
				mockService.On("GetByID", mock.Anything, mockUserUUID).Return(mockUser, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   string(expectedUserJSON),
		},
		{
			name:        "Invalid User ID Format",
			userIDParam: "not-a-uuid",
			setupMock:   func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid user ID format"}`,
		},
		{
			name:        "User Not Found",
			userIDParam: uuid.New().String(), // Use a different valid UUID
			setupMock: func(mockService *MockUserService) {
				mockService.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"User not found"}`,
		},
		{
			name:        "Internal Server Error",
			userIDParam: uuid.New().String(),
			setupMock: func(mockService *MockUserService) {
				mockService.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get user"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tc.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			rr := httptest.NewRecorder()
			_, router := gin.CreateTestContext(rr)
			router.GET("/users/:id", handler.GetUserByID)

			req, _ := http.NewRequest(http.MethodGet, "/users/"+tc.userIDParam, nil)
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
