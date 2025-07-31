package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"seta-training/internal/models"
	"seta-training/pkg/auth"
)

// MockUserRepository is a mock implementation of UserRepositoryInterface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetAll() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) UsernameExists(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

// MockJWTManager is a mock implementation of JWTManagerInterface
type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateToken(user *models.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) ValidateToken(tokenString string) (*auth.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *MockJWTManager) RefreshToken(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

func TestUserService_CreateUser(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTManager)
	service := NewUserService(mockRepo, mockJWT)

	input := &CreateUserInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     models.RoleManager,
	}

	// Mock expectations
	mockRepo.On("EmailExists", input.Email).Return(false, nil)
	mockRepo.On("UsernameExists", input.Username).Return(false, nil)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	// Test
	user, err := service.CreateUser(input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, input.Username, user.Username)
	assert.Equal(t, input.Email, user.Email)
	assert.Equal(t, input.Role, user.Role)
	assert.NotEmpty(t, user.PasswordHash)
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_EmailExists(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTManager)
	service := NewUserService(mockRepo, mockJWT)

	input := &CreateUserInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     models.RoleManager,
	}

	// Mock expectations
	mockRepo.On("EmailExists", input.Email).Return(true, nil)

	// Test
	user, err := service.CreateUser(input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "email already exists")
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTManager)
	service := NewUserService(mockRepo, mockJWT)

	hashedPassword, _ := auth.HashPassword("password123")
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         models.RoleManager,
	}

	input := &LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock expectations
	mockRepo.On("GetByEmail", input.Email).Return(user, nil)
	mockJWT.On("GenerateToken", user).Return("mock-jwt-token", nil)

	// Test
	response, err := service.Login(input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, user, response.User)
	assert.Equal(t, "mock-jwt-token", response.Token)
	mockRepo.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestUserService_Login_InvalidPassword(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTManager)
	service := NewUserService(mockRepo, mockJWT)

	hashedPassword, _ := auth.HashPassword("correctpassword")
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         models.RoleManager,
	}

	input := &LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	// Mock expectations
	mockRepo.On("GetByEmail", input.Email).Return(user, nil)

	// Test
	response, err := service.Login(input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid email or password")
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetAllUsers(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTManager)
	service := NewUserService(mockRepo, mockJWT)

	expectedUsers := []models.User{
		{
			ID:       uuid.New(),
			Username: "user1",
			Email:    "user1@example.com",
			Role:     models.RoleManager,
		},
		{
			ID:       uuid.New(),
			Username: "user2",
			Email:    "user2@example.com",
			Role:     models.RoleMember,
		},
	}

	// Mock expectations
	mockRepo.On("GetAll").Return(expectedUsers, nil)

	// Test
	users, err := service.GetAllUsers()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
	mockRepo.AssertExpectations(t)
}
