package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/go-user-service/internal/user/dto"
	"github.com/example/go-user-service/internal/user/model"
	"github.com/example/go-user-service/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function to get a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// MockUserService is a mock implementation of UserService.
type MockUserService struct {
	mock.Mock
}

// RegisterUser mocks the RegisterUser method.
func (m *MockUserService) RegisterUser(req *dto.UserRegisterRequest) (*model.User, error) {
	args := m.Called(req)
	user, ok := args.Get(0).(*model.User)
	if !ok && args.Get(0) != nil {
		panic("Mocked RegisterUser must return *model.User or nil")
	}
	return user, args.Error(1)
}

// GetUserByID mocks the GetUserByID method.
func (m *MockUserService) GetUserByID(id uint) (*model.User, error) {
	args := m.Called(id)
	user, ok := args.Get(0).(*model.User)
	if !ok && args.Get(0) != nil {
		panic("Mocked GetUserByID must return *model.User or nil")
	}
	return user, args.Error(1)
}

// GetUserByEmail mocks the GetUserByEmail method.
func (m *MockUserService) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	user, ok := args.Get(0).(*model.User)
	if !ok && args.Get(0) != nil {
		panic("Mocked GetUserByEmail must return *model.User or nil")
	}
	return user, args.Error(1)
}

// UpdateUser mocks the UpdateUser method.
func (m *MockUserService) UpdateUser(id uint, req *dto.UserUpdateRequest) (*model.User, error) {
	args := m.Called(id, req)
	user, ok := args.Get(0).(*model.User)
	if !ok && args.Get(0) != nil {
		panic("Mocked UpdateUser must return *model.User or nil")
	}
	return user, args.Error(1)
}

// DeleteUser mocks the DeleteUser method.
func (m *MockUserService) DeleteUser(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// SetupRouter sets up a Gin router for testing.
func SetupRouter(handler UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/users", handler.Register)
	r.GET("/users/:id", handler.GetUserByID)
	r.GET("/users", handler.GetUserByEmail)
	r.PUT("/users/:id", handler.UpdateUser)
	r.DELETE("/users/:id", handler.DeleteUser)
	return r
}

// Helper function to create a test context and recorder
func CreateTestContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.Request = req

	return c, w
}

// TestRegisterHandler tests the Register handler.
func TestRegisterHandler(t *testing.T) {
	var resp response.Response
	var err error
	// Test case 1: Successful registration
	dt := time.Now()
	expectedUser := &model.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: dt,
		UpdatedAt: dt,
	}

	mockService := new(MockUserService)
	// Expect RegisterUser to be called with the request and return the expected user and no error
	mockService.On("RegisterUser", &dto.UserRegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}).Return(expectedUser, nil).Once()

	handler := NewUserHandler(mockService)
	r := SetupRouter(handler)

	registerReq := dto.UserRegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}

	// Create a test context and recorder
	w := httptest.NewRecorder()
	reqBody, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "Success", resp.Message)

	// Unmarshal the data field into a UserResponse DTO
	var userResp dto.UserResponse
	dataBytes, _ := json.Marshal(resp.Data)
	err = json.Unmarshal(dataBytes, &userResp)
	assert.NoError(t, err)

	// Assert the user data in the response
	assert.Equal(t, expectedUser.ID, userResp.ID)
	assert.Equal(t, expectedUser.Username, userResp.Username)
	assert.Equal(t, expectedUser.Email, userResp.Email)
	// Note: Comparing time.Time directly can be tricky, compare relevant fields or use a tolerance if needed
	// assert.WithinDuration(t, expectedUser.CreatedAt, userResp.CreatedAt, time.Second)
	// assert.WithinDuration(t, expectedUser.UpdatedAt, userResp.UpdatedAt, time.Second)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 2: Invalid request body
	mockService = new(MockUserService) // Create a new mock for the next test case
	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	// Create a test context with invalid JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEqual(t, http.StatusOK, resp.Code)
	assert.Equal(t, "Invalid request body", resp.Message)

	// Assert that the mock service method was NOT called
	mockService.AssertExpectations(t)

	// Test case 3: Service returns an error
	mockService = new(MockUserService) // Create a new mock for the next test case
	// Expect RegisterUser to be called and return an error
	mockService.On("RegisterUser", &dto.UserRegisterRequest{
		Username: "erroruser",
		Password: "password456",
		Email:    "error@example.com",
	}).Return(nil, errors.New("service error")).Once()

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	registerReq = dto.UserRegisterRequest{
		Username: "erroruser",
		Password: "password456",
		Email:    "error@example.com",
	}

	// Create a test context and recorder
	w = httptest.NewRecorder()
	reqBody, _ = json.Marshal(registerReq)
	req, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEqual(t, http.StatusOK, resp.Code)
	assert.Equal(t, "Failed to register user", resp.Message)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)
}

// TestGetUserByIDHandler tests the GetUserByID handler.
func TestGetUserByIDHandler(t *testing.T) {
	var resp response.Response
	var err error
	// Test case 1: Successful retrieval
	dt := time.Now()
	expectedUser := &model.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: dt,
		UpdatedAt: dt,
	}

	mockService := new(MockUserService)
	// Expect GetUserByID to be called with ID 1 and return the expected user and no error
	mockService.On("GetUserByID", uint(1)).Return(expectedUser, nil).Once()

	handler := NewUserHandler(mockService)
	r := SetupRouter(handler)

	// Create a test context and recorder for GET request with ID parameter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/1", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "Success", resp.Message)

	// Unmarshal the data field into a UserResponse DTO
	var userResp dto.UserResponse
	dataBytes, _ := json.Marshal(resp.Data)
	err = json.Unmarshal(dataBytes, &userResp)
	assert.NoError(t, err)

	// Assert the user data in the response
	assert.Equal(t, expectedUser.ID, userResp.ID)
	assert.Equal(t, expectedUser.Username, userResp.Username)
	assert.Equal(t, expectedUser.Email, userResp.Email)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 2: User not found
	mockService = new(MockUserService) // Create a new mock for the next test case
	// Expect GetUserByID to be called with ID 2 and return nil and no error (user not found scenario)
	mockService.On("GetUserByID", uint(2)).Return(nil, nil).Once()

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	// Create a test context and recorder for GET request with ID parameter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/users/2", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusNotFound, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Equal(t, "User not found", resp.Message)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 3: Service returns an error
	mockService = new(MockUserService) // Create a new mock for the next test case
	// Expect GetUserByID to be called with ID 3 and return an error
	mockService.On("GetUserByID", uint(3)).Return(nil, errors.New("service error")).Once()

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	// Create a test context and recorder for GET request with ID parameter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/users/3", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "Failed to get user", resp.Message)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 4: Invalid ID parameter
	mockService = new(MockUserService) // Create a new mock for the next test case
	// No service method is expected to be called for invalid input

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	// Create a test context and recorder for GET request with invalid ID parameter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/users/invalid", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, "Invalid user ID", resp.Message)

	// Assert that the mock service method was NOT called
	mockService.AssertExpectations(t)
}

// TestGetUserByEmailHandler tests the GetUserByEmail handler.
func TestGetUserByEmailHandler(t *testing.T) {
	var resp response.Response
	var err error
	// Test case 1: Successful retrieval
	dt := time.Now()
	expectedUser := &model.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: dt,
		UpdatedAt: dt,
	}

	mockService := new(MockUserService)
	// Expect GetUserByEmail to be called with the email and return the expected user and no error
	mockService.On("GetUserByEmail", "test@example.com").Return(expectedUser, nil).Once()

	handler := NewUserHandler(mockService)
	r := SetupRouter(handler)

	// Create a test context and recorder for GET request with email query parameter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users?email=test@example.com", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "Success", resp.Message)

	// Unmarshal the data field into a UserResponse DTO
	var userResp dto.UserResponse
	dataBytes, _ := json.Marshal(resp.Data)
	err = json.Unmarshal(dataBytes, &userResp)
	assert.NoError(t, err)

	// Assert the user data in the response
	assert.Equal(t, expectedUser.ID, userResp.ID)
	assert.Equal(t, expectedUser.Username, userResp.Username)
	assert.Equal(t, expectedUser.Email, userResp.Email)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 2: User not found
	mockService = new(MockUserService) // Create a new mock for the next test case
	// Expect GetUserByEmail to be called with a different email and return nil and no error
	mockService.On("GetUserByEmail", "notfound@example.com").Return(nil, nil).Once()

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	// Create a test context and recorder for GET request with email query parameter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/users?email=notfound@example.com", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusNotFound, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Equal(t, "User not found", resp.Message)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 3: Service returns an error
	mockService = new(MockUserService) // Create a new mock for the next test case
	// Expect GetUserByEmail to be called and return an error
	mockService.On("GetUserByEmail", "error@example.com").Return(nil, errors.New("service error")).Once()

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	// Create a test context and recorder for GET request with email query parameter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/users?email=error@example.com", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "Failed to get user", resp.Message)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 4: Missing email query parameter
	mockService = new(MockUserService) // Create a new mock for the next test case
	// No service method is expected to be called for missing parameter

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	// Create a test context and recorder for GET request without email query parameter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/users", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, "Email query parameter is required", resp.Message)

	// Assert that the mock service method was NOT called
	mockService.AssertExpectations(t)
}

// TestUpdateUserHandler tests the UpdateUser handler.
func TestUpdateUserHandler(t *testing.T) {
	var resp response.Response
	var err error
	// Test case 1: Successful update
	dt := time.Now()
	expectedUser := &model.User{
		ID:        1,
		Username:  "updateduser",
		Email:     "updated@example.com",
		CreatedAt: dt,
		UpdatedAt: dt,
	}

	updateReq := &dto.UserUpdateRequest{
		Username: stringPtr("updateduser"),
		Email:    stringPtr("updated@example.com"),
	}

	mockService := new(MockUserService)
	// Expect UpdateUser to be called with ID 1 and the request, return the updated user and no error
	mockService.On("UpdateUser", uint(1), updateReq).Return(expectedUser, nil).Once()

	handler := NewUserHandler(mockService)
	r := SetupRouter(handler)

	// Create a test context and recorder for PUT request with ID parameter and body
	w := httptest.NewRecorder()
	reqBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "Success", resp.Message)

	// Unmarshal the data field into a UserResponse DTO
	var userResp dto.UserResponse
	dataBytes, _ := json.Marshal(resp.Data)
	err = json.Unmarshal(dataBytes, &userResp)
	assert.NoError(t, err)

	// Assert the user data in the response
	assert.Equal(t, expectedUser.ID, userResp.ID)
	assert.Equal(t, expectedUser.Username, userResp.Username)
	assert.Equal(t, expectedUser.Email, userResp.Email)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 2: Invalid request body
	mockService = new(MockUserService) // Create a new mock for the next test case
	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	// Create a test context with invalid JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/users/1", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEqual(t, http.StatusOK, resp.Code)
	assert.Equal(t, "Invalid request body", resp.Message)

	// Assert that the mock service method was NOT called
	mockService.AssertExpectations(t)

	// Test case 3: User not found
	mockService = new(MockUserService) // Create a new mock for the next test case
	// Expect UpdateUser to be called with ID 2 and the request, return nil and an error (user not found scenario)
	mockService.On("UpdateUser", uint(2), &dto.UserUpdateRequest{
		Username: stringPtr("notfounduser"),
		Email:    stringPtr("notfound@example.com"),
	}).Return(nil, errors.New("user not found")).Once()

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	updateReqNotFound := &dto.UserUpdateRequest{
		Username: stringPtr("notfounduser"),
		Email:    stringPtr("notfound@example.com"),
	}

	// Create a test context and recorder for PUT request with ID parameter and body
	w = httptest.NewRecorder()
	reqBody, _ = json.Marshal(updateReqNotFound)
	req, _ = http.NewRequest(http.MethodPut, "/users/2", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusInternalServerError, w.Code) // Assuming service error maps to 500

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "Failed to update user", resp.Message)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 4: Service returns a generic error
	mockService = new(MockUserService) // Create a new mock for the next test case
	// Expect UpdateUser to be called with ID 3 and the request, return nil and an error
	mockService.On("UpdateUser", uint(3), &dto.UserUpdateRequest{
		Username: stringPtr("erroruser"),
		Email:    stringPtr("error@example.com"),
	}).Return(nil, errors.New("service error")).Once()

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	updateReqError := &dto.UserUpdateRequest{
		Username: stringPtr("erroruser"),
		Email:    stringPtr("error@example.com"),
	}

	// Create a test context and recorder for PUT request with ID parameter and body
	w = httptest.NewRecorder()
	reqBody, _ = json.Marshal(updateReqError)
	req, _ = http.NewRequest(http.MethodPut, "/users/3", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "Failed to update user", resp.Message)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 5: Invalid ID parameter
	mockService = new(MockUserService) // Create a new mock for the next test case
	// No service method is expected to be called for invalid input

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	updateReqInvalidID := &dto.UserUpdateRequest{
		Username: stringPtr("invalididuser"),
		Email:    stringPtr("invalidid@example.com"),
	}

	// Create a test context and recorder for PUT request with invalid ID parameter and body
	w = httptest.NewRecorder()
	reqBody, _ = json.Marshal(updateReqInvalidID)
	req, _ = http.NewRequest(http.MethodPut, "/users/invalid", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, "Invalid user ID", resp.Message)

	// Assert that the mock service method was NOT called
	mockService.AssertExpectations(t)
}

// TODO: Add tests for DeleteUser handlers

// TestDeleteUserHandler tests the DeleteUser handler.
func TestDeleteUserHandler(t *testing.T) {
	var resp response.Response
	var err error
	// Test case 1: Successful deletion
	mockService := new(MockUserService)
	// Expect DeleteUser to be called with ID 1, return no error
	mockService.On("DeleteUser", uint(1)).Return(nil).Once()

	handler := NewUserHandler(mockService)
	r := SetupRouter(handler)

	// Create a test context and recorder for DELETE request with ID parameter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/users/1", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "Success", resp.Message)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 2: User not found
	mockService = new(MockUserService) // Create a new mock for the next test case
	// Expect DeleteUser to be called with ID 2, return an error (user not found scenario)
	mockService.On("DeleteUser", uint(2)).Return(errors.New("user not found")).Once()

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	// Create a test context and recorder for DELETE request with ID parameter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/users/2", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusInternalServerError, w.Code) // Assuming service error maps to 500

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "Failed to delete user", resp.Message)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 3: Service returns a generic error
	mockService = new(MockUserService) // Create a new mock for the next test case
	// Expect DeleteUser to be called with ID 3, return an error
	mockService.On("DeleteUser", uint(3)).Return(errors.New("service error")).Once()

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	// Create a test context and recorder for DELETE request with ID parameter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/users/3", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "Failed to delete user", resp.Message)

	// Assert that the mock service method was called as expected
	mockService.AssertExpectations(t)

	// Test case 4: Invalid ID parameter
	mockService = new(MockUserService) // Create a new mock for the next test case
	// No service method is expected to be called for invalid input

	handler = NewUserHandler(mockService)
	r = SetupRouter(handler)

	// Create a test context and recorder for DELETE request with invalid ID parameter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/users/invalid", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	resp = response.Response{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, "Invalid user ID", resp.Message)

	// Assert that the mock service method was NOT called
	mockService.AssertExpectations(t)
}
