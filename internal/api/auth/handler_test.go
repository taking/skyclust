package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"skyclust/internal/api/common"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(req domain.CreateUserRequest) (*domain.User, string, error) {
	args := m.Called(req)
	return args.Get(0).(*domain.User), args.String(1), args.Error(2)
}

func (m *MockAuthService) Login(username, password string) (*domain.User, string, error) {
	args := m.Called(username, password)
	return args.Get(0).(*domain.User), args.String(1), args.Error(2)
}

func (m *MockAuthService) LoginWithContext(username, password, clientIP, userAgent string) (*domain.User, string, error) {
	args := m.Called(username, password, clientIP, userAgent)
	return args.Get(0).(*domain.User), args.String(1), args.Error(2)
}

func (m *MockAuthService) ValidateToken(token string) (*domain.User, error) {
	args := m.Called(token)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) Logout(userID uuid.UUID, token string) error {
	args := m.Called(userID, token)
	return args.Error(0)
}

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetUsers(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.User, int64, error) {
	args := m.Called(ctx, limit, offset, filters)
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetUsersWithFilters(filters domain.UserFilters) ([]*domain.User, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) UpdateUserDirect(user *domain.User) (*domain.User, error) {
	args := m.Called(user)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) DeleteUserByID(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) GetUserStats() (*domain.UserStats, error) {
	args := m.Called()
	return args.Get(0).(*domain.UserStats), args.Error(1)
}

func (m *MockUserService) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

// TestRegister tests the Register handler
func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService, *MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful registration",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
			},
			mockSetup: func(authService *MockAuthService, userService *MockUserService) {
				user := &domain.User{
					ID:        uuid.New(),
					Username:  "testuser",
					Email:     "test@example.com",
					IsActive:  true,
					CreatedAt: time.Now(),
				}
				authService.On("Register", mock.AnythingOfType("domain.CreateUserRequest")).Return(user, "jwt-token", nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"username": "testuser",
				// Missing email and password
			},
			mockSetup: func(authService *MockAuthService, userService *MockUserService) {
				// No mock setup needed for invalid request - should not call service
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "registration failure",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
			},
			mockSetup: func(authService *MockAuthService, userService *MockUserService) {
				authService.On("Register", mock.AnythingOfType("domain.CreateUserRequest")).Return((*domain.User)(nil), "", assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to register user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAuthService := &MockAuthService{}
			mockUserService := &MockUserService{}
			tt.mockSetup(mockAuthService, mockUserService)

			// Create handler
			handler := NewHandler(mockAuthService, mockUserService)

			// Create test context
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = req

			// Execute
			handler.Register(ginContext)

			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedError != "" {
				var response common.APIResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.expectedError)
			}

			// Verify mocks
			mockAuthService.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
		})
	}
}

// TestLogin tests the Login handler
func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService, *MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful login",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "password123",
			},
			mockSetup: func(authService *MockAuthService, userService *MockUserService) {
				user := &domain.User{
					ID:       uuid.New(),
					Username: "testuser",
					Email:    "test@example.com",
					IsActive: true,
				}
				authService.On("Login", "testuser", "password123").Return(user, "jwt-token", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "wrongpassword",
			},
			mockSetup: func(authService *MockAuthService, userService *MockUserService) {
				authService.On("Login", "testuser", "wrongpassword").Return((*domain.User)(nil), "", domain.NewDomainError(domain.ErrCodeInvalidCredentials, "Invalid credentials", 401))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAuthService := &MockAuthService{}
			mockUserService := &MockUserService{}
			tt.mockSetup(mockAuthService, mockUserService)

			// Create handler
			handler := NewHandler(mockAuthService, mockUserService)

			// Create test context
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = req

			// Execute
			handler.Login(ginContext)

			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Verify mocks
			mockAuthService.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
		})
	}
}

// TestMe tests the Me handler
func TestMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockAuthService, *MockUserService)
		expectedStatus int
	}{
		{
			name:   "successful user info retrieval",
			userID: "user123",
			mockSetup: func(authService *MockAuthService, userService *MockUserService) {
				user := &domain.User{
					ID:       uuid.MustParse("user123"),
					Username: "testuser",
					Email:    "test@example.com",
					IsActive: true,
				}
				userService.On("GetUser", mock.AnythingOfType("*context.valueCtx"), "user123").Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			mockSetup: func(authService *MockAuthService, userService *MockUserService) {
				userService.On("GetUser", mock.AnythingOfType("*context.valueCtx"), "nonexistent").Return((*domain.User)(nil), domain.NewDomainError(domain.ErrCodeNotFound, "User not found", 404))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAuthService := &MockAuthService{}
			mockUserService := &MockUserService{}
			tt.mockSetup(mockAuthService, mockUserService)

			// Create handler
			handler := NewHandler(mockAuthService, mockUserService)

			// Create test context
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = req
			ginContext.Set("user_id", tt.userID)

			// Execute
			handler.Me(ginContext)

			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Verify mocks
			mockAuthService.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
		})
	}
}

// TestGetUsers tests the GetUsers handler
func TestGetUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockAuthService, *MockUserService)
		expectedStatus int
	}{
		{
			name:        "successful user list retrieval",
			queryParams: "?limit=10&offset=0",
			mockSetup: func(authService *MockAuthService, userService *MockUserService) {
				users := []*domain.User{
					{ID: uuid.New(), Username: "user1", Email: "user1@example.com"},
					{ID: uuid.New(), Username: "user2", Email: "user2@example.com"},
				}
				userService.On("GetUsers", mock.AnythingOfType("*context.valueCtx"), 10, 0, mock.AnythingOfType("map[string]interface {}")).Return(users, int64(2), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid pagination parameters",
			queryParams: "?limit=invalid&offset=invalid",
			mockSetup: func(authService *MockAuthService, userService *MockUserService) {
				// No mock setup needed for invalid parameters
			},
			expectedStatus: http.StatusOK, // Should use default values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAuthService := &MockAuthService{}
			mockUserService := &MockUserService{}
			tt.mockSetup(mockAuthService, mockUserService)

			// Create handler
			handler := NewHandler(mockAuthService, mockUserService)

			// Create test context
			req := httptest.NewRequest(http.MethodGet, "/users"+tt.queryParams, nil)
			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = req
			ginContext.Set("user_id", "admin")

			// Execute
			handler.GetUsers(ginContext)

			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Verify mocks
			mockAuthService.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
		})
	}
}
