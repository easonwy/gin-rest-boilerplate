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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
	realServiceUser "github.com/yi-tech/go-user-service/internal/service/user" // Import for sentinel errors
)

// stringPtr is a helper function to get a pointer to a string. Useful for DTOs with optional string fields.
func stringPtr(s string) *string {
	return &s
}

// MockUserService is a mock type for the UserService interface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, input domainUser.RegisterUserInput) (*domainUser.User, error) {
	args := m.Called(ctx, input)
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

func (m *MockUserService) Update(ctx context.Context, id uuid.UUID, params domainUser.UpdateUserParams) (*domainUser.User, error) {
	args := m.Called(ctx, id, params)
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

// createMockDomainUser creates a mock domainUser.User object for testing.
func createMockDomainUser(id uuid.UUID, email, firstName, lastName string) *domainUser.User {
	now := time.Now()
	return &domainUser.User{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  "hashedpassword",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestNewUserHandler(t *testing.T) {
	service := new(MockUserService)
	handler := NewHandler(service, zaptest.NewLogger(t))
	assert.NotNil(t, handler)
	assert.Equal(t, service, handler.userService)
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
		Password:  baseUser.Password,  // Password not changed
		CreatedAt: baseUser.CreatedAt, // CreatedAt does not change
		UpdatedAt: time.Now(),         // UpdatedAt is set by the service
	}

	// Expected JSON structure is now checked directly in the test case

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
				// Mock GetByID to return the user
				mockService.On("GetByID", mock.Anything, mockUserUUID).Return(baseUser, nil).Once()
				// Mock Update to return the updated user
				mockService.On("Update", mock.Anything, mockUserUUID, mock.MatchedBy(func(params domainUser.UpdateUserParams) bool {
					return params.FirstName == updatedFirstName && params.LastName == updatedLastName
				})).Return(successUserForMockReturn, nil).Once()
			},
			expectedStatus: http.StatusOK,
			// expectedBody will be compared as a map, see assertion logic below
			ignoreBody: true, // Special handling for UpdatedAt
		},
		{
			name:           "Invalid User ID Format",
			userIDParam:    "not-a-uuid",
			requestBody:    UserUpdateRequest{FirstName: stringPtr("Test"), LastName: stringPtr("User")},
			setupMock:      func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"code":400,"message":"Invalid user ID format"}`,
		},
		{
			name:           "Invalid Request Data - Malformed JSON",
			userIDParam:    mockUserUUID.String(),
			requestBody:    "not json",
			setupMock:      func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"code":400,"message":"Invalid request data"}`,
		},
		// {
		// 	name:        "Invalid Request Data - Missing FirstName",
		// 	userIDParam: mockUserUUID.String(),
		// 	requestBody: UserUpdateRequest{LastName: stringPtr("User")}, // Missing FirstName
		// 	setupMock: func(mockService *MockUserService) {
		// 		// This test is commented out because the handler's UpdateProfile method
		// 		// was intentionally changed to allow partial updates (meaning FirstName
		// 		// or LastName can be omitted in the request without causing a direct
		// 		// validation error in the handler). The service layer now handles
		// 		// the logic of applying provided fields. If FirstName were truly mandatory
		// 		// at the HTTP layer, this test would be valid, but the DTO uses *string,
		// 		// implying optionality.
		// 		mockUser := createMockDomainUser(mockUserUUID, "test@example.com", "Test", "User")
		// 		mockService.On("GetByID", mock.Anything, mockUserUUID).Return(mockUser, nil).Once()
		// 		// If partial update with empty FirstName is valid and service is called:
		// 		// updatedUser := *mockUser // copy
		// 		// updatedUser.LastName = "User" // LastName is updated
		// 		// mockService.On("Update", mock.Anything, mockUserUUID, mock.MatchedBy(func(params domainUser.UpdateUserParams) bool {
		// 		// 	return params.FirstName == "" && params.LastName == "User" // FirstName would be its zero value if not provided
		// 		// })).Return(&updatedUser, nil).Once()
		// 	},
		// 	expectedStatus: http.StatusBadRequest, // This would change if partial update is allowed and successful
		// 	expectedBody:   `{"code":400,"message":"first name is required"}`, // This would change
		// },
		{
			name:        "User Not Found",
			userIDParam: "00000000-0000-0000-0000-000000000001", // Fixed UUID for consistent testing
			requestBody: UserUpdateRequest{FirstName: stringPtr("Test"), LastName: stringPtr("User")},
			setupMock: func(mockService *MockUserService) {
				// Use the same UUID as in userIDParam
				userUUID, err := uuid.Parse("00000000-0000-0000-0000-000000000001")
				assert.NoError(t, err)
				mockService.On("GetByID", mock.Anything, userUUID).Return(nil, realServiceUser.ErrUserNotFound).Once()
				// Update should not be called, so no mock for Update in this specific path.
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"code":404,"message":"user not found"}`, // Message from realServiceUser.ErrUserNotFound.Error()
		},
		{
			name:        "Internal Server Error",
			userIDParam: mockUserUUID.String(),
			requestBody: UserUpdateRequest{FirstName: stringPtr("Test"), LastName: stringPtr("User")},
			setupMock: func(mockService *MockUserService) {
				errUser := createMockDomainUser(mockUserUUID, "test@example.com", "Test", "User")
				mockService.On("GetByID", mock.Anything, mockUserUUID).Return(errUser, nil).Once()
				mockService.On("Update", mock.Anything, mockUserUUID, mock.MatchedBy(func(params domainUser.UpdateUserParams) bool {
					return params.FirstName == "Test" && params.LastName == "User"
				})).Return(nil, errors.New("internal error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"code":500,"message":"Something went wrong. Please try again later."}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tc.setupMock(mockService)

			handler := NewHandler(mockService, logger)

			rr := httptest.NewRecorder()
			_, router := gin.CreateTestContext(rr)
			router.PUT("/users/:id", handler.UpdateProfile) // Changed from UpdateUser to UpdateProfile

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

			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.ignoreBody {
				// For success case, parse the response and check the fields
				var responseBody map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
				assert.NoError(t, err)

				// Check the response structure
				assert.Equal(t, float64(http.StatusOK), responseBody["code"])
				assert.Equal(t, "Success", responseBody["message"])

				// Get the data object
				data, ok := responseBody["data"].(map[string]interface{})
				assert.True(t, ok, "data should be present in response")

				// Check that updated_at is present and not empty
				updatedAt, ok := data["updated_at"].(string)
				assert.True(t, ok, "updated_at should be present in response data")
				assert.NotEmpty(t, updatedAt, "updated_at should not be empty")

				// Remove updated_at for comparison
				delete(data, "updated_at")

				// Check the fields we care about
				assert.Equal(t, mockUserUUID.String(), data["id"])
				assert.Equal(t, "original@example.com", data["email"])
				assert.Equal(t, "UpdatedFirst", data["first_name"])
				assert.Equal(t, "UpdatedLast", data["last_name"])
				// Just verify created_at is a valid RFC3339 timestamp
				_, err = time.Parse(time.RFC3339, data["created_at"].(string))
				assert.NoError(t, err, "created_at should be a valid RFC3339 timestamp")
			} else {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}

			// Assert that all expected mock calls were made
			mockService.AssertExpectations(t)
		})
	}
}
