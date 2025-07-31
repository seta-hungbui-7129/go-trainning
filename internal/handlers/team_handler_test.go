package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"seta-training/internal/models"
	"seta-training/internal/services"
	"seta-training/pkg/auth"
)

// MockTeamService is a mock implementation of TeamServiceInterface
type MockTeamService struct {
	mock.Mock
}

func (m *MockTeamService) CreateTeam(input *services.CreateTeamInput, creatorID uuid.UUID) (*models.Team, error) {
	args := m.Called(input, creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamService) AddMember(teamID, userID, managerID uuid.UUID) error {
	args := m.Called(teamID, userID, managerID)
	return args.Error(0)
}

func (m *MockTeamService) RemoveMember(teamID, userID, managerID uuid.UUID) error {
	args := m.Called(teamID, userID, managerID)
	return args.Error(0)
}

func (m *MockTeamService) AddManager(teamID, userID, requestorID uuid.UUID) error {
	args := m.Called(teamID, userID, requestorID)
	return args.Error(0)
}

func (m *MockTeamService) RemoveManager(teamID, userID, requestorID uuid.UUID) error {
	args := m.Called(teamID, userID, requestorID)
	return args.Error(0)
}

func (m *MockTeamService) GetTeam(teamID uuid.UUID) (*models.Team, error) {
	args := m.Called(teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamService) GetAllTeams() ([]models.Team, error) {
	args := m.Called()
	return args.Get(0).([]models.Team), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func setupAuthContext(c *gin.Context, userID uuid.UUID, role models.UserRole) {
	claims := &auth.Claims{
		UserID: userID,
		Role:   role,
	}
	c.Set("claims", claims)
}

func TestTeamHandler_CreateTeam_Success(t *testing.T) {
	// Setup
	mockService := new(MockTeamService)
	handler := NewTeamHandler(mockService)
	router := setupTestRouter()

	userID := uuid.New()
	expectedTeam := &models.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}

	input := services.CreateTeamInput{
		Name:     "Test Team",
		Managers: []services.TeamMemberInput{},
		Members:  []services.TeamMemberInput{},
	}

	// Mock expectations
	mockService.On("CreateTeam", &input, userID).Return(expectedTeam, nil)

	// Setup route with auth context
	router.POST("/teams", func(c *gin.Context) {
		setupAuthContext(c, userID, models.RoleManager)
		handler.CreateTeam(c)
	})

	// Prepare request
	jsonData, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Test
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.Team
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTeam.Name, response.Name)
	mockService.AssertExpectations(t)
}

func TestTeamHandler_CreateTeam_InvalidInput(t *testing.T) {
	// Setup
	mockService := new(MockTeamService)
	handler := NewTeamHandler(mockService)
	router := setupTestRouter()

	userID := uuid.New()

	// Setup route with auth context
	router.POST("/teams", func(c *gin.Context) {
		setupAuthContext(c, userID, models.RoleManager)
		handler.CreateTeam(c)
	})

	// Prepare request with invalid JSON
	req, _ := http.NewRequest("POST", "/teams", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Test
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid input")
}

func TestTeamHandler_GetTeam_Success(t *testing.T) {
	// Setup
	mockService := new(MockTeamService)
	handler := NewTeamHandler(mockService)
	router := setupTestRouter()

	teamID := uuid.New()
	userID := uuid.New()
	expectedTeam := &models.Team{
		ID:   teamID,
		Name: "Test Team",
	}

	// Mock expectations
	mockService.On("GetTeam", teamID).Return(expectedTeam, nil)

	// Setup route with auth context
	router.GET("/teams/:teamId", func(c *gin.Context) {
		setupAuthContext(c, userID, models.RoleManager)
		handler.GetTeam(c)
	})

	// Prepare request
	req, _ := http.NewRequest("GET", "/teams/"+teamID.String(), nil)
	w := httptest.NewRecorder()

	// Test
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.Team
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTeam.Name, response.Name)
	mockService.AssertExpectations(t)
}

func TestTeamHandler_GetTeam_InvalidID(t *testing.T) {
	// Setup
	mockService := new(MockTeamService)
	handler := NewTeamHandler(mockService)
	router := setupTestRouter()

	userID := uuid.New()

	// Setup route with auth context
	router.GET("/teams/:teamId", func(c *gin.Context) {
		setupAuthContext(c, userID, models.RoleManager)
		handler.GetTeam(c)
	})

	// Prepare request with invalid UUID
	req, _ := http.NewRequest("GET", "/teams/invalid-uuid", nil)
	w := httptest.NewRecorder()

	// Test
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid team ID")
}

func TestTeamHandler_AddMember_Success(t *testing.T) {
	// Setup
	mockService := new(MockTeamService)
	handler := NewTeamHandler(mockService)
	router := setupTestRouter()

	teamID := uuid.New()
	userID := uuid.New()
	managerID := uuid.New()

	input := struct {
		UserID uuid.UUID `json:"userId"`
	}{
		UserID: userID,
	}

	// Mock expectations
	mockService.On("AddMember", teamID, userID, managerID).Return(nil)

	// Setup route with auth context
	router.POST("/teams/:teamId/members", func(c *gin.Context) {
		setupAuthContext(c, managerID, models.RoleManager)
		handler.AddMember(c)
	})

	// Prepare request
	jsonData, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/teams/"+teamID.String()+"/members", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Test
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Member added successfully", response["message"])
	mockService.AssertExpectations(t)
}
