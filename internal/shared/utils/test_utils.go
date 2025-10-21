package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/gorm"
	"skyclust/internal/domain"
	"skyclust/internal/shared/responses"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestContext provides test context utilities
type TestContext struct {
	ginContext *gin.Context
	recorder   *httptest.ResponseRecorder
	request    *http.Request
}

// NewTestContext creates a new test context
func NewTestContext(method, url string, body interface{}) *TestContext {
	gin.SetMode(gin.TestMode)

	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = req

	return &TestContext{
		ginContext: ginContext,
		recorder:   recorder,
		request:    req,
	}
}

// GetGinContext returns the gin context
func (tc *TestContext) GetGinContext() *gin.Context {
	return tc.ginContext
}

// GetResponse returns the response recorder
func (tc *TestContext) GetResponse() *httptest.ResponseRecorder {
	return tc.recorder
}

// SetUserID sets the user ID in the context
func (tc *TestContext) SetUserID(userID string) {
	tc.ginContext.Set("user_id", userID)
}

// SetUserRole sets the user role in the context
func (tc *TestContext) SetUserRole(role string) {
	tc.ginContext.Set("user_role", role)
}

// SetRequestID sets the request ID in the context
func (tc *TestContext) SetRequestID(requestID string) {
	tc.ginContext.Set("request_id", requestID)
}

// AssertStatus asserts the HTTP status code
func (tc *TestContext) AssertStatus(t *testing.T, expectedStatus int) {
	assert.Equal(t, expectedStatus, tc.recorder.Code)
}

// AssertJSON asserts the JSON response
func (tc *TestContext) AssertJSON(t *testing.T, expected interface{}) {
	var actual, expectedJSON interface{}

	err := json.Unmarshal(tc.recorder.Body.Bytes(), &actual)
	assert.NoError(t, err)

	expectedBytes, err := json.Marshal(expected)
	assert.NoError(t, err)

	err = json.Unmarshal(expectedBytes, &expectedJSON)
	assert.NoError(t, err)

	assert.Equal(t, expectedJSON, actual)
}

// AssertContains asserts that the response contains a string
func (tc *TestContext) AssertContains(t *testing.T, expected string) {
	assert.Contains(t, tc.recorder.Body.String(), expected)
}

// AssertHeader asserts a response header
func (tc *TestContext) AssertHeader(t *testing.T, key, expected string) {
	assert.Equal(t, expected, tc.recorder.Header().Get(key))
}

// MockService provides a mock service for testing
type MockService struct {
	mock.Mock
}

// MockUserService provides a mock user service
type MockUserService struct {
	mock.Mock
}

// MockWorkspaceService provides a mock workspace service
type MockWorkspaceService struct {
	mock.Mock
}

// MockNotificationService provides a mock notification service
type MockNotificationService struct {
	mock.Mock
}

// MockAuthService provides a mock auth service
type MockAuthService struct {
	mock.Mock
}

// TestData provides test data utilities
type TestData struct{}

// NewTestData creates a new test data instance
func NewTestData() *TestData {
	return &TestData{}
}

// GetTestUser returns a test user
func (td *TestData) GetTestUser() *domain.User {
	return &domain.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetTestWorkspace returns a test workspace
func (td *TestData) GetTestWorkspace() *domain.Workspace {
	return &domain.Workspace{
		ID:          uuid.New().String(),
		Name:        "Test Workspace",
		Description: "Test workspace description",
		OwnerID:     uuid.New().String(),
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// GetTestNotification returns a test notification
func (td *TestData) GetTestNotification() *domain.Notification {
	return &domain.Notification{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Title:     "Test Notification",
		Message:   "Test notification message",
		Type:      "info",
		Priority:  "medium",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
}

// GetTestCredentials returns test credentials
func (td *TestData) GetTestCredentials() *domain.Credential {
	return &domain.Credential{
		ID:            uuid.New(),
		UserID:        uuid.New(),
		Provider:      "aws",
		Name:          "Test Credentials",
		EncryptedData: []byte("encrypted_test_data"),
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// TestHelper provides test helper functions
type TestHelper struct{}

// NewTestHelper creates a new test helper
func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

// CreateTestRequest creates a test HTTP request
func (th *TestHelper) CreateTestRequest(method, url string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// CreateTestResponse creates a test HTTP response recorder
func (th *TestHelper) CreateTestResponse() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// CreateTestGinContext creates a test gin context
func (th *TestHelper) CreateTestGinContext(req *http.Request, recorder *httptest.ResponseRecorder) *gin.Context {
	gin.SetMode(gin.TestMode)
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = req
	return ginContext
}

// AssertAPIResponse asserts an API response structure
func (th *TestHelper) AssertAPIResponse(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int, expectedSuccess bool) {
	assert.Equal(t, expectedStatus, response.Code)

	var apiResponse responses.APIResponse
	err := json.Unmarshal(response.Body.Bytes(), &apiResponse)
	assert.NoError(t, err)
	assert.Equal(t, expectedSuccess, apiResponse.Success)
}

// AssertErrorResponse asserts an error response
func (th *TestHelper) AssertErrorResponse(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int, expectedCode string) {
	assert.Equal(t, expectedStatus, response.Code)

	var apiResponse responses.APIResponse
	err := json.Unmarshal(response.Body.Bytes(), &apiResponse)
	assert.NoError(t, err)
	assert.False(t, apiResponse.Success)
	assert.Equal(t, expectedCode, apiResponse.Code)
}

// AssertSuccessResponse asserts a success response
func (th *TestHelper) AssertSuccessResponse(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(t, expectedStatus, response.Code)

	var apiResponse responses.APIResponse
	err := json.Unmarshal(response.Body.Bytes(), &apiResponse)
	assert.NoError(t, err)
	assert.True(t, apiResponse.Success)
}

// TestDatabase provides test database utilities
type TestDatabase struct {
	db *gorm.DB
}

// NewTestDatabase creates a new test database
func NewTestDatabase(db *gorm.DB) *TestDatabase {
	return &TestDatabase{
		db: db,
	}
}

// SetupTestData sets up test data
func (tdb *TestDatabase) SetupTestData() error {
	// Create test users
	users := []domain.User{
		{ID: uuid.MustParse("user1"), Username: "testuser1", Email: "test1@example.com", Active: true},
		{ID: uuid.MustParse("user2"), Username: "testuser2", Email: "test2@example.com", Active: true},
	}

	for _, user := range users {
		if err := tdb.db.Create(&user).Error; err != nil {
			return err
		}
	}

	// Create test workspaces
	workspaces := []domain.Workspace{
		{ID: "workspace1", Name: "Test Workspace 1", OwnerID: "user1", Active: true},
		{ID: "workspace2", Name: "Test Workspace 2", OwnerID: "user2", Active: true},
	}

	for _, workspace := range workspaces {
		if err := tdb.db.Create(&workspace).Error; err != nil {
			return err
		}
	}

	return nil
}

// CleanupTestData cleans up test data
func (tdb *TestDatabase) CleanupTestData() error {
	// Clean up in reverse order to avoid foreign key constraints
	if err := tdb.db.Where("id IN ?", []string{"workspace1", "workspace2"}).Delete(&domain.Workspace{}).Error; err != nil {
		return err
	}

	if err := tdb.db.Where("id IN ?", []string{"user1", "user2"}).Delete(&domain.User{}).Error; err != nil {
		return err
	}

	return nil
}

// TestMetrics provides test metrics
type TestMetrics struct {
	RequestCount    int64
	ErrorCount      int64
	AverageResponse float64
	SlowQueries     int64
}

// NewTestMetrics creates new test metrics
func NewTestMetrics() *TestMetrics {
	return &TestMetrics{}
}

// RecordRequest records a test request
func (tm *TestMetrics) RecordRequest(duration time.Duration, isError bool) {
	tm.RequestCount++
	if isError {
		tm.ErrorCount++
	}
	if duration > 100*time.Millisecond {
		tm.SlowQueries++
	}
	tm.AverageResponse = float64(tm.RequestCount) / float64(duration.Milliseconds())
}

// GetMetrics returns current metrics
func (tm *TestMetrics) GetMetrics() *TestMetrics {
	return tm
}
