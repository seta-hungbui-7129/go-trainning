package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"seta-training/internal/models"
	"seta-training/pkg/auth"
)

// SetupTestRouter creates a test Gin router
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// SetupAuthContext sets up authentication context for testing
func SetupAuthContext(c *gin.Context, userID uuid.UUID, role models.UserRole) {
	claims := &auth.Claims{
		UserID: userID,
		Role:   role,
	}
	c.Set("claims", claims)
}

// CreateTestUser creates a test user
func CreateTestUser(role models.UserRole) *models.User {
	return &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     role,
	}
}

// CreateTestTeam creates a test team
func CreateTestTeam() *models.Team {
	return &models.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
}

// CreateTestFolder creates a test folder
func CreateTestFolder(ownerID uuid.UUID) *models.Folder {
	return &models.Folder{
		ID:      uuid.New(),
		Name:    "Test Folder",
		OwnerID: ownerID,
	}
}

// CreateTestNote creates a test note
func CreateTestNote(folderID, ownerID uuid.UUID) *models.Note {
	return &models.Note{
		ID:       uuid.New(),
		Title:    "Test Note",
		Body:     "Test note content",
		FolderID: folderID,
		OwnerID:  ownerID,
	}
}

// MakeJSONRequest creates a JSON HTTP request
func MakeJSONRequest(method, url string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}
	
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// AssertJSONResponse asserts that the response contains expected JSON
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedBody interface{}) {
	assert.Equal(t, expectedStatus, w.Code)
	
	if expectedBody != nil {
		var actualBody interface{}
		err := json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.NoError(t, err)
		
		expectedJSON, _ := json.Marshal(expectedBody)
		actualJSON, _ := json.Marshal(actualBody)
		assert.JSONEq(t, string(expectedJSON), string(actualJSON))
	}
}

// AssertErrorResponse asserts that the response contains an error message
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedErrorContains string) {
	assert.Equal(t, expectedStatus, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	errorMsg, exists := response["error"]
	assert.True(t, exists, "Response should contain error field")
	assert.Contains(t, errorMsg.(string), expectedErrorContains)
}

// AssertSuccessResponse asserts that the response contains a success message
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMessageContains string) {
	assert.Equal(t, expectedStatus, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	message, exists := response["message"]
	assert.True(t, exists, "Response should contain message field")
	assert.Contains(t, message.(string), expectedMessageContains)
}

// TestDatabase provides utilities for database testing
type TestDatabase struct {
	// Add database testing utilities here if needed
}

// MockAuthMiddleware creates a mock authentication middleware for testing
func MockAuthMiddleware(userID uuid.UUID, role models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetupAuthContext(c, userID, role)
		c.Next()
	}
}

// MockManagerMiddleware creates a mock manager middleware for testing
func MockManagerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In real middleware, this would check if user is manager
		// For testing, we assume the user is already set as manager in context
		c.Next()
	}
}

// TestConfig provides test configuration
type TestConfig struct {
	DatabaseURL string
	JWTSecret   string
}

// GetTestConfig returns test configuration
func GetTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseURL: "postgres://test:test@localhost:5432/test_db?sslmode=disable",
		JWTSecret:   "test-secret-key",
	}
}
