package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"seta-training/internal/models"
)

// MockTeamRepository is a mock implementation of TeamRepositoryInterface
type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(team *models.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetByID(id uuid.UUID) (*models.Team, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetAll() ([]models.Team, error) {
	args := m.Called()
	return args.Get(0).([]models.Team), args.Error(1)
}

func (m *MockTeamRepository) AddManager(teamID, userID uuid.UUID) error {
	args := m.Called(teamID, userID)
	return args.Error(0)
}

func (m *MockTeamRepository) RemoveManager(teamID, userID uuid.UUID) error {
	args := m.Called(teamID, userID)
	return args.Error(0)
}

func (m *MockTeamRepository) AddMember(teamID, userID uuid.UUID) error {
	args := m.Called(teamID, userID)
	return args.Error(0)
}

func (m *MockTeamRepository) RemoveMember(teamID, userID uuid.UUID) error {
	args := m.Called(teamID, userID)
	return args.Error(0)
}

func (m *MockTeamRepository) IsManager(teamID, userID uuid.UUID) (bool, error) {
	args := m.Called(teamID, userID)
	return args.Bool(0), args.Error(1)
}

func TestTeamService_CreateTeam_Success(t *testing.T) {
	// Setup
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo)

	creatorID := uuid.New()
	creator := &models.User{
		ID:   creatorID,
		Role: models.RoleManager,
	}

	input := &CreateTeamInput{
		Name:     "Test Team",
		Managers: []TeamMemberInput{},
		Members:  []TeamMemberInput{},
	}

	expectedTeam := &models.Team{
		ID:   uuid.New(),
		Name: input.Name,
		Managers: []models.User{*creator},
		Members:  []models.User{},
	}

	// Mock expectations
	mockUserRepo.On("GetByID", creatorID).Return(creator, nil)
	mockTeamRepo.On("Create", mock.AnythingOfType("*models.Team")).Return(nil)
	mockTeamRepo.On("AddManager", mock.AnythingOfType("uuid.UUID"), creatorID).Return(nil)
	mockTeamRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(expectedTeam, nil)

	// Test
	team, err := service.CreateTeam(input, creatorID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, input.Name, team.Name)
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTeamService_CreateTeam_NonManagerCreator(t *testing.T) {
	// Setup
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo)

	creatorID := uuid.New()
	creator := &models.User{
		ID:   creatorID,
		Role: models.RoleMember, // Not a manager
	}

	input := &CreateTeamInput{
		Name:     "Test Team",
		Managers: []TeamMemberInput{},
		Members:  []TeamMemberInput{},
	}

	// Mock expectations
	mockUserRepo.On("GetByID", creatorID).Return(creator, nil)

	// Test
	team, err := service.CreateTeam(input, creatorID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, team)
	assert.Contains(t, err.Error(), "only managers can create teams")
	mockUserRepo.AssertExpectations(t)
}

func TestTeamService_AddMember_Success(t *testing.T) {
	// Setup
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo)

	teamID := uuid.New()
	userID := uuid.New()
	managerID := uuid.New()

	user := &models.User{
		ID:   userID,
		Role: models.RoleMember,
	}

	// Mock expectations
	mockTeamRepo.On("IsManager", teamID, managerID).Return(true, nil)
	mockUserRepo.On("GetByID", userID).Return(user, nil)
	mockTeamRepo.On("AddMember", teamID, userID).Return(nil)

	// Test
	err := service.AddMember(teamID, userID, managerID)

	// Assert
	assert.NoError(t, err)
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTeamService_AddMember_NotManager(t *testing.T) {
	// Setup
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo)

	teamID := uuid.New()
	userID := uuid.New()
	managerID := uuid.New()

	// Mock expectations
	mockTeamRepo.On("IsManager", teamID, managerID).Return(false, nil)

	// Test
	err := service.AddMember(teamID, userID, managerID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient permissions")
	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_AddManager_Success(t *testing.T) {
	// Setup
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo)

	teamID := uuid.New()
	userID := uuid.New()
	requestorID := uuid.New()

	user := &models.User{
		ID:   userID,
		Role: models.RoleManager,
	}

	// Mock expectations
	mockTeamRepo.On("IsManager", teamID, requestorID).Return(true, nil)
	mockUserRepo.On("GetByID", userID).Return(user, nil)
	mockTeamRepo.On("AddManager", teamID, userID).Return(nil)

	// Test
	err := service.AddManager(teamID, userID, requestorID)

	// Assert
	assert.NoError(t, err)
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTeamService_AddManager_UserNotManager(t *testing.T) {
	// Setup
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo)

	teamID := uuid.New()
	userID := uuid.New()
	requestorID := uuid.New()

	user := &models.User{
		ID:   userID,
		Role: models.RoleMember, // Not a manager
	}

	// Mock expectations
	mockTeamRepo.On("IsManager", teamID, requestorID).Return(true, nil)
	mockUserRepo.On("GetByID", userID).Return(user, nil)

	// Test
	err := service.AddManager(teamID, userID, requestorID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user must be a manager")
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTeamService_GetTeam(t *testing.T) {
	// Setup
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo)

	teamID := uuid.New()
	expectedTeam := &models.Team{
		ID:   teamID,
		Name: "Test Team",
	}

	// Mock expectations
	mockTeamRepo.On("GetByID", teamID).Return(expectedTeam, nil)

	// Test
	team, err := service.GetTeam(teamID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedTeam, team)
	mockTeamRepo.AssertExpectations(t)
}
