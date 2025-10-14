package workspace

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

// MockWorkspaceService is a mock implementation of WorkspaceService
type MockWorkspaceService struct {
	mock.Mock
}

func (m *MockWorkspaceService) CreateWorkspace(ctx context.Context, req domain.CreateWorkspaceRequest) (*domain.Workspace, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceService) GetWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceService) UpdateWorkspace(ctx context.Context, id string, req domain.UpdateWorkspaceRequest) (*domain.Workspace, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWorkspaceService) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceService) AddUserToWorkspace(ctx context.Context, workspaceID, userID string) error {
	args := m.Called(ctx, workspaceID, userID)
	return args.Error(0)
}

func (m *MockWorkspaceService) RemoveUserFromWorkspace(ctx context.Context, workspaceID, userID string) error {
	args := m.Called(ctx, workspaceID, userID)
	return args.Error(0)
}

func (m *MockWorkspaceService) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]*domain.User, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]*domain.User), args.Error(1)
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

// TestCreateWorkspace tests the CreateWorkspace handler
func TestCreateWorkspace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		userID         string
		mockSetup      func(*MockWorkspaceService, *MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful workspace creation",
			requestBody: map[string]interface{}{
				"name":        "Test Workspace",
				"description": "Test workspace description",
			},
			userID: "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				workspace := &domain.Workspace{
					ID:          uuid.New().String(),
					Name:        "Test Workspace",
					Description: "Test workspace description",
					OwnerID:     "user123",
					IsActive:    true,
					CreatedAt:   time.Now(),
				}
				workspaceService.On("CreateWorkspace", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("domain.CreateWorkspaceRequest")).Return(workspace, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"name": "", // Empty name should fail validation
			},
			userID: "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				// No mock setup needed for invalid request
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "workspace creation failure",
			requestBody: map[string]interface{}{
				"name":        "Test Workspace",
				"description": "Test workspace description",
			},
			userID: "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				workspaceService.On("CreateWorkspace", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("domain.CreateWorkspaceRequest")).Return((*domain.Workspace)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to create workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockWorkspaceService := &MockWorkspaceService{}
			mockUserService := &MockUserService{}
			tt.mockSetup(mockWorkspaceService, mockUserService)

			// Create handler
			handler := NewHandler(mockWorkspaceService, mockUserService)

			// Create test context
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/workspaces", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = req
			ginContext.Set("user_id", tt.userID)

			// Execute
			handler.CreateWorkspace(ginContext)

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
			mockWorkspaceService.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
		})
	}
}

// TestGetWorkspaces tests the GetWorkspaces handler
func TestGetWorkspaces(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		userID         string
		mockSetup      func(*MockWorkspaceService, *MockUserService)
		expectedStatus int
	}{
		{
			name:        "successful workspace list retrieval",
			queryParams: "?limit=10&offset=0",
			userID:      "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				workspaces := []*domain.Workspace{
					{ID: "workspace1", Name: "Workspace 1", OwnerID: "user123"},
					{ID: "workspace2", Name: "Workspace 2", OwnerID: "user123"},
				}
				workspaceService.On("GetUserWorkspaces", mock.AnythingOfType("*context.valueCtx"), "user123").Return(workspaces, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid pagination parameters",
			queryParams: "?limit=invalid&offset=invalid",
			userID:      "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				workspaces := []*domain.Workspace{}
				workspaceService.On("GetUserWorkspaces", mock.AnythingOfType("*context.valueCtx"), "user123").Return(workspaces, nil)
			},
			expectedStatus: http.StatusOK, // Should use default values
		},
		{
			name:        "workspace retrieval failure",
			queryParams: "?limit=10&offset=0",
			userID:      "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				workspaceService.On("GetUserWorkspaces", mock.AnythingOfType("*context.valueCtx"), "user123").Return(([]*domain.Workspace)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockWorkspaceService := &MockWorkspaceService{}
			mockUserService := &MockUserService{}
			tt.mockSetup(mockWorkspaceService, mockUserService)

			// Create handler
			handler := NewHandler(mockWorkspaceService, mockUserService)

			// Create test context
			req := httptest.NewRequest(http.MethodGet, "/workspaces"+tt.queryParams, nil)
			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = req
			ginContext.Set("user_id", tt.userID)

			// Execute
			handler.GetWorkspaces(ginContext)

			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Verify mocks
			mockWorkspaceService.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
		})
	}
}

// TestGetWorkspace tests the GetWorkspace handler
func TestGetWorkspace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		workspaceID    string
		userID         string
		mockSetup      func(*MockWorkspaceService, *MockUserService)
		expectedStatus int
	}{
		{
			name:        "successful workspace retrieval",
			workspaceID: "workspace123",
			userID:      "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				workspace := &domain.Workspace{
					ID:       "workspace123",
					Name:     "Test Workspace",
					OwnerID:  "user123",
					IsActive: true,
				}
				workspaceService.On("GetWorkspace", mock.AnythingOfType("*context.valueCtx"), "workspace123").Return(workspace, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "workspace not found",
			workspaceID: "nonexistent",
			userID:      "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				workspaceService.On("GetWorkspace", mock.AnythingOfType("*context.valueCtx"), "nonexistent").Return((*domain.Workspace)(nil), domain.NewDomainError(domain.ErrCodeNotFound, "Workspace not found", 404))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockWorkspaceService := &MockWorkspaceService{}
			mockUserService := &MockUserService{}
			tt.mockSetup(mockWorkspaceService, mockUserService)

			// Create handler
			handler := NewHandler(mockWorkspaceService, mockUserService)

			// Create test context
			req := httptest.NewRequest(http.MethodGet, "/workspaces/"+tt.workspaceID, nil)
			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = req
			ginContext.Set("user_id", tt.userID)
			ginContext.Params = []gin.Param{{Key: "id", Value: tt.workspaceID}}

			// Execute
			handler.GetWorkspace(ginContext)

			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Verify mocks
			mockWorkspaceService.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
		})
	}
}

// TestUpdateWorkspace tests the UpdateWorkspace handler
func TestUpdateWorkspace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		workspaceID    string
		requestBody    interface{}
		userID         string
		mockSetup      func(*MockWorkspaceService, *MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "successful workspace update",
			workspaceID: "workspace123",
			requestBody: map[string]interface{}{
				"name":        "Updated Workspace",
				"description": "Updated description",
			},
			userID: "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				workspace := &domain.Workspace{
					ID:          "workspace123",
					Name:        "Updated Workspace",
					Description: "Updated description",
					OwnerID:     "user123",
					IsActive:    true,
				}
				workspaceService.On("UpdateWorkspace", mock.AnythingOfType("*context.valueCtx"), "workspace123", mock.AnythingOfType("domain.UpdateWorkspaceRequest")).Return(workspace, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "workspace not found",
			workspaceID: "nonexistent",
			requestBody: map[string]interface{}{
				"name": "Updated Workspace",
			},
			userID: "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				workspaceService.On("UpdateWorkspace", mock.AnythingOfType("*context.valueCtx"), "nonexistent", mock.AnythingOfType("domain.UpdateWorkspaceRequest")).Return((*domain.Workspace)(nil), domain.NewDomainError(domain.ErrCodeNotFound, "Workspace not found", 404))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockWorkspaceService := &MockWorkspaceService{}
			mockUserService := &MockUserService{}
			tt.mockSetup(mockWorkspaceService, mockUserService)

			// Create handler
			handler := NewHandler(mockWorkspaceService, mockUserService)

			// Create test context
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/workspaces/"+tt.workspaceID, bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = req
			ginContext.Set("user_id", tt.userID)
			ginContext.Params = []gin.Param{{Key: "id", Value: tt.workspaceID}}

			// Execute
			handler.UpdateWorkspace(ginContext)

			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Verify mocks
			mockWorkspaceService.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
		})
	}
}

// TestDeleteWorkspace tests the DeleteWorkspace handler
func TestDeleteWorkspace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		workspaceID    string
		userID         string
		mockSetup      func(*MockWorkspaceService, *MockUserService)
		expectedStatus int
	}{
		{
			name:        "successful workspace deletion",
			workspaceID: "workspace123",
			userID:      "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				workspaceService.On("DeleteWorkspace", mock.AnythingOfType("*context.valueCtx"), "workspace123").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "workspace not found",
			workspaceID: "nonexistent",
			userID:      "user123",
			mockSetup: func(workspaceService *MockWorkspaceService, userService *MockUserService) {
				workspaceService.On("DeleteWorkspace", mock.AnythingOfType("*context.valueCtx"), "nonexistent").Return(domain.NewDomainError(domain.ErrCodeNotFound, "Workspace not found", 404))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockWorkspaceService := &MockWorkspaceService{}
			mockUserService := &MockUserService{}
			tt.mockSetup(mockWorkspaceService, mockUserService)

			// Create handler
			handler := NewHandler(mockWorkspaceService, mockUserService)

			// Create test context
			req := httptest.NewRequest(http.MethodDelete, "/workspaces/"+tt.workspaceID, nil)
			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = req
			ginContext.Set("user_id", tt.userID)
			ginContext.Params = []gin.Param{{Key: "id", Value: tt.workspaceID}}

			// Execute
			handler.DeleteWorkspace(ginContext)

			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Verify mocks
			mockWorkspaceService.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
		})
	}
}
