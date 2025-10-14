package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"skyclust/internal/domain"
)

// MockAuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(req domain.CreateUserRequest) (*domain.User, string, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*domain.User), args.String(1), args.Error(2)
}

func (m *MockAuthService) Login(username, password string) (*domain.User, string, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*domain.User), args.String(1), args.Error(2)
}

func (m *MockAuthService) ValidateToken(token string) (*domain.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) LoginWithContext(username, password, clientIP, userAgent string) (*domain.User, string, error) {
	args := m.Called(username, password, clientIP, userAgent)
	if args.Get(0) == nil {
		return nil, "", args.Error(3)
	}
	return args.Get(0).(*domain.User), args.String(1), args.Error(3)
}

func (m *MockAuthService) Logout(userID uuid.UUID, token string) error {
	args := m.Called(userID, token)
	return args.Error(0)
}

// MockUserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetUsers(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.User, int64, error) {
	args := m.Called(ctx, limit, offset, filters)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) UpdateUser(ctx context.Context, username string, req domain.UpdateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, username, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, username string) error {
	args := m.Called(ctx, username)
	return args.Error(0)
}

func (m *MockUserService) Authenticate(ctx context.Context, username, password string) (*domain.User, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	args := m.Called(ctx, username, oldPassword, newPassword)
	return args.Error(0)
}

// Admin-specific methods
func (m *MockUserService) GetUserByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetUsersWithFilters(filters domain.UserFilters) ([]*domain.User, int64, error) {
	args := m.Called(filters)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) UpdateUserDirect(user *domain.User) (*domain.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) DeleteUserByID(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) GetUserStats() (*domain.UserStats, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserStats), args.Error(1)
}

func TestAuthHandler_Register_Simple(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockAuthService, *MockUserService)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful registration",
			requestBody: gin.H{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
			},
			setupMocks: func(authService *MockAuthService, userService *MockUserService) {
				user := &domain.User{
					ID:       uuid.New(),
					Username: "testuser",
					Email:    "test@example.com",
				}
				authService.On("Register", mock.AnythingOfType("domain.CreateUserRequest")).Return(user, "jwt-token", nil)
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			authService := new(MockAuthService)
			userService := new(MockUserService)
			tt.setupMocks(authService, userService)

			// Create handler
			handler := NewAuthHandler(authService, userService)

			// Setup router
			router := gin.New()
			router.POST("/register", handler.Register)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.NotNil(t, response["data"])
			}

			// Verify mocks
			authService.AssertExpectations(t)
			userService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockAuthService, *MockUserService)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful login",
			requestBody: gin.H{
				"username": "testuser",
				"password": "password123",
			},
			setupMocks: func(authService *MockAuthService, userService *MockUserService) {
				user := &domain.User{
					ID:       uuid.New(),
					Username: "testuser",
					Email:    "test@example.com",
				}
				authService.On("Login", "testuser", "password123").Return(user, "jwt-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "invalid credentials",
			requestBody: gin.H{
				"username": "testuser",
				"password": "wrongpassword",
			},
			setupMocks: func(authService *MockAuthService, userService *MockUserService) {
				authService.On("Login", "testuser", "wrongpassword").Return(nil, "", assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name: "invalid request body",
			requestBody: gin.H{
				"username": "testuser",
				// Missing password
			},
			setupMocks: func(authService *MockAuthService, userService *MockUserService) {
				// No mock setup needed for invalid request
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			authService := new(MockAuthService)
			userService := new(MockUserService)
			tt.setupMocks(authService, userService)

			// Create handler
			handler := NewAuthHandler(authService, userService)

			// Setup router
			router := gin.New()
			router.POST("/login", handler.Login)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.NotNil(t, response["data"])
			}

			// Verify mocks
			authService.AssertExpectations(t)
			userService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_GetUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func(*MockAuthService, *MockUserService)
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "successful users retrieval",
			queryParams: "?page=1&limit=10",
			setupMocks: func(authService *MockAuthService, userService *MockUserService) {
				users := []*domain.User{
					{
						ID:       uuid.New(),
						Username: "user1",
						Email:    "user1@example.com",
					},
					{
						ID:       uuid.New(),
						Username: "user2",
						Email:    "user2@example.com",
					},
				}
				userService.On("GetUsers", mock.AnythingOfType("*context.valueCtx"), 10, 0, mock.AnythingOfType("map[string]interface {}")).Return(users, int64(2), nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "invalid page parameter",
			queryParams: "?page=invalid&limit=10",
			setupMocks: func(authService *MockAuthService, userService *MockUserService) {
				// No mock setup needed for invalid parameter
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:        "invalid limit parameter",
			queryParams: "?page=1&limit=invalid",
			setupMocks: func(authService *MockAuthService, userService *MockUserService) {
				// No mock setup needed for invalid parameter
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			authService := new(MockAuthService)
			userService := new(MockUserService)
			tt.setupMocks(authService, userService)

			// Create handler
			handler := NewAuthHandler(authService, userService)

			// Setup router
			router := gin.New()
			router.GET("/users", handler.GetUsers)

			// Create request
			req := httptest.NewRequest("GET", "/users"+tt.queryParams, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.NotNil(t, response["data"])
			}

			// Verify mocks
			authService.AssertExpectations(t)
			userService.AssertExpectations(t)
		})
	}
}
