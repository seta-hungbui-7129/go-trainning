package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"seta-training/internal/models"
	"seta-training/pkg/auth"
	"seta-training/pkg/logger"
)

// MockUserService is a mock implementation of UserServiceInterface for import testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(input *CreateUserInput) (*models.User, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Login(input *LoginInput) (*LoginResponse, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LoginResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetAllUsers() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserService) ValidateToken(tokenString string) (*auth.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

// MockImportLogger is a mock logger for testing
type MockImportLogger struct {
	mock.Mock
}

func (m *MockImportLogger) Debug(msg string, fields ...logger.Field) {
	// Allow any debug calls without strict expectations
}

func (m *MockImportLogger) Info(msg string, fields ...logger.Field) {
	// Allow any info calls without strict expectations
}

func (m *MockImportLogger) Warn(msg string, fields ...logger.Field) {
	// Allow any warn calls without strict expectations
}

func (m *MockImportLogger) Error(msg string, fields ...logger.Field) {
	// Allow any error calls without strict expectations
}

func (m *MockImportLogger) Fatal(msg string, fields ...logger.Field) {
	// Allow any fatal calls without strict expectations
}

func (m *MockImportLogger) WithContext(ctx context.Context) logger.Logger {
	return m
}

func (m *MockImportLogger) WithFields(fields ...logger.Field) logger.Logger {
	return m
}

func TestImportService_ImportUsersFromCSV_Success(t *testing.T) {
	// Setup
	mockUserService := new(MockUserService)
	mockLogger := new(MockImportLogger)
	service := NewImportService(mockUserService, mockLogger)

	// CSV data with multiple users
	csvData := `username,email,password,role
john.doe,john.doe@example.com,password123,manager
jane.smith,jane.smith@example.com,password456,member
bob.wilson,bob.wilson@example.com,password789,member`

	// Mock logger allows any calls without expectations

	// Mock user creation - all succeed
	mockUserService.On("CreateUser", mock.MatchedBy(func(input *CreateUserInput) bool {
		return input.Username == "john.doe"
	})).Return(&models.User{
		ID:       uuid.New(),
		Username: "john.doe",
		Email:    "john.doe@example.com",
		Role:     models.RoleManager,
	}, nil)

	mockUserService.On("CreateUser", mock.MatchedBy(func(input *CreateUserInput) bool {
		return input.Username == "jane.smith"
	})).Return(&models.User{
		ID:       uuid.New(),
		Username: "jane.smith",
		Email:    "jane.smith@example.com",
		Role:     models.RoleMember,
	}, nil)

	mockUserService.On("CreateUser", mock.MatchedBy(func(input *CreateUserInput) bool {
		return input.Username == "bob.wilson"
	})).Return(&models.User{
		ID:       uuid.New(),
		Username: "bob.wilson",
		Email:    "bob.wilson@example.com",
		Role:     models.RoleMember,
	}, nil)

	// Test configuration
	config := ImportConfig{
		WorkerCount:    2,
		BatchSize:      10,
		Timeout:        10 * time.Second,
		MaxRecords:     100,
		SkipDuplicates: true,
	}

	// Test
	ctx := context.Background()
	summary, err := service.ImportUsersFromCSV(ctx, strings.NewReader(csvData), config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 3, summary.TotalRecords)
	assert.Equal(t, 3, summary.SuccessCount)
	assert.Equal(t, 0, summary.FailureCount)
	assert.Len(t, summary.Results, 3)

	// Verify all results are successful
	for _, result := range summary.Results {
		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
		assert.NotEmpty(t, result.UserID)
	}

	mockUserService.AssertExpectations(t)
}

func TestImportService_ImportUsersFromCSV_PartialFailure(t *testing.T) {
	// Setup
	mockUserService := new(MockUserService)
	mockLogger := new(MockImportLogger)
	service := NewImportService(mockUserService, mockLogger)

	// CSV data with one invalid role
	csvData := `username,email,password,role
john.doe,john.doe@example.com,password123,manager
jane.smith,jane.smith@example.com,password456,invalid_role
bob.wilson,bob.wilson@example.com,password789,member`

	// Mock logger allows any calls without expectations

	// Mock user creation - first and third succeed
	mockUserService.On("CreateUser", mock.MatchedBy(func(input *CreateUserInput) bool {
		return input.Username == "john.doe"
	})).Return(&models.User{
		ID:       uuid.New(),
		Username: "john.doe",
		Email:    "john.doe@example.com",
		Role:     models.RoleManager,
	}, nil)

	mockUserService.On("CreateUser", mock.MatchedBy(func(input *CreateUserInput) bool {
		return input.Username == "bob.wilson"
	})).Return(&models.User{
		ID:       uuid.New(),
		Username: "bob.wilson",
		Email:    "bob.wilson@example.com",
		Role:     models.RoleMember,
	}, nil)

	// Test configuration
	config := ImportConfig{
		WorkerCount:    2,
		BatchSize:      10,
		Timeout:        10 * time.Second,
		MaxRecords:     100,
		SkipDuplicates: true,
	}

	// Test
	ctx := context.Background()
	summary, err := service.ImportUsersFromCSV(ctx, strings.NewReader(csvData), config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 3, summary.TotalRecords)
	assert.Equal(t, 2, summary.SuccessCount)
	assert.Equal(t, 1, summary.FailureCount)
	assert.Len(t, summary.Results, 3)

	// Check that one result failed due to invalid role
	failedCount := 0
	for _, result := range summary.Results {
		if !result.Success {
			failedCount++
			assert.Contains(t, result.Error, "invalid role")
		}
	}
	assert.Equal(t, 1, failedCount)

	mockUserService.AssertExpectations(t)
}

func TestImportService_ImportUsersFromCSV_InvalidHeader(t *testing.T) {
	// Setup
	mockUserService := new(MockUserService)
	mockLogger := new(MockImportLogger)
	service := NewImportService(mockUserService, mockLogger)

	// CSV data with invalid header
	csvData := `name,email,pass,type
john.doe,john.doe@example.com,password123,manager`

	// Test configuration
	config := DefaultImportConfig()

	// Test
	ctx := context.Background()
	summary, err := service.ImportUsersFromCSV(ctx, strings.NewReader(csvData), config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, summary)
	assert.Contains(t, err.Error(), "invalid CSV header")
}

func TestImportService_ImportUsersFromCSV_EmptyFile(t *testing.T) {
	// Setup
	mockUserService := new(MockUserService)
	mockLogger := new(MockImportLogger)
	service := NewImportService(mockUserService, mockLogger)

	// CSV data with only header
	csvData := `username,email,password,role`

	// Mock logger allows any calls without expectations

	// Test configuration
	config := DefaultImportConfig()

	// Test
	ctx := context.Background()
	summary, err := service.ImportUsersFromCSV(ctx, strings.NewReader(csvData), config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 0, summary.TotalRecords)
	assert.Equal(t, 0, summary.SuccessCount)
	assert.Equal(t, 0, summary.FailureCount)
	assert.Len(t, summary.Results, 0)
}

func TestImportService_ImportUsersFromCSV_MaxRecordsLimit(t *testing.T) {
	// Setup
	mockUserService := new(MockUserService)
	mockLogger := new(MockImportLogger)
	service := NewImportService(mockUserService, mockLogger)

	// CSV data with 3 users
	csvData := `username,email,password,role
john.doe,john.doe@example.com,password123,manager
jane.smith,jane.smith@example.com,password456,member
bob.wilson,bob.wilson@example.com,password789,member`

	// Mock logger allows any calls without expectations

	// Mock user creation for first 2 users only
	mockUserService.On("CreateUser", mock.MatchedBy(func(input *CreateUserInput) bool {
		return input.Username == "john.doe"
	})).Return(&models.User{
		ID:       uuid.New(),
		Username: "john.doe",
		Email:    "john.doe@example.com",
		Role:     models.RoleManager,
	}, nil)

	mockUserService.On("CreateUser", mock.MatchedBy(func(input *CreateUserInput) bool {
		return input.Username == "jane.smith"
	})).Return(&models.User{
		ID:       uuid.New(),
		Username: "jane.smith",
		Email:    "jane.smith@example.com",
		Role:     models.RoleMember,
	}, nil)

	// Test configuration with max 2 records
	config := ImportConfig{
		WorkerCount:    2,
		BatchSize:      10,
		Timeout:        10 * time.Second,
		MaxRecords:     2, // Limit to 2 records
		SkipDuplicates: true,
	}

	// Test
	ctx := context.Background()
	summary, err := service.ImportUsersFromCSV(ctx, strings.NewReader(csvData), config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 2, summary.TotalRecords) // Should only process 2 records
	assert.Equal(t, 2, summary.SuccessCount)
	assert.Equal(t, 0, summary.FailureCount)
	assert.Len(t, summary.Results, 2)

	mockUserService.AssertExpectations(t)
}
