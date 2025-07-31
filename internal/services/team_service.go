package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"seta-training/internal/models"
	"seta-training/internal/repositories"
)

type TeamService struct {
	teamRepo repositories.TeamRepositoryInterface
	userRepo repositories.UserRepositoryInterface
}

func NewTeamService(teamRepo repositories.TeamRepositoryInterface, userRepo repositories.UserRepositoryInterface) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

type CreateTeamInput struct {
	Name     string                `json:"teamName" binding:"required,min=3,max=100"`
	Managers []TeamMemberInput     `json:"managers"`
	Members  []TeamMemberInput     `json:"members"`
}

type TeamMemberInput struct {
	ID   uuid.UUID `json:"managerId,omitempty"`
	Name string    `json:"managerName,omitempty"`
}

func (s *TeamService) CreateTeam(input *CreateTeamInput, creatorID uuid.UUID) (*models.Team, error) {
	// Verify creator is a manager
	creator, err := s.userRepo.GetByID(creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}
	if !creator.IsManager() {
		return nil, errors.New("only managers can create teams")
	}

	// Create team
	team := &models.Team{
		Name: input.Name,
	}

	if err := s.teamRepo.Create(team); err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Add creator as manager
	if err := s.teamRepo.AddManager(team.ID, creatorID); err != nil {
		return nil, fmt.Errorf("failed to add creator as manager: %w", err)
	}

	// Add additional managers
	for _, manager := range input.Managers {
		if manager.ID != creatorID { // Don't add creator twice
			// Verify user exists and is a manager
			user, err := s.userRepo.GetByID(manager.ID)
			if err != nil {
				continue // Skip invalid users
			}
			if user.IsManager() {
				s.teamRepo.AddManager(team.ID, manager.ID)
			}
		}
	}

	// Add members
	for _, member := range input.Members {
		// Verify user exists
		if _, err := s.userRepo.GetByID(member.ID); err == nil {
			s.teamRepo.AddMember(team.ID, member.ID)
		}
	}

	// Return team with relationships loaded
	return s.teamRepo.GetByID(team.ID)
}

func (s *TeamService) AddMember(teamID, userID, managerID uuid.UUID) error {
	// Verify manager has permission
	if err := s.verifyManagerPermission(teamID, managerID); err != nil {
		return err
	}

	// Verify user exists
	if _, err := s.userRepo.GetByID(userID); err != nil {
		return errors.New("user not found")
	}

	return s.teamRepo.AddMember(teamID, userID)
}

func (s *TeamService) RemoveMember(teamID, userID, managerID uuid.UUID) error {
	// Verify manager has permission
	if err := s.verifyManagerPermission(teamID, managerID); err != nil {
		return err
	}

	return s.teamRepo.RemoveMember(teamID, userID)
}

func (s *TeamService) AddManager(teamID, userID, requestorID uuid.UUID) error {
	// Verify requestor has permission
	if err := s.verifyManagerPermission(teamID, requestorID); err != nil {
		return err
	}

	// Verify user exists and is a manager
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}
	if !user.IsManager() {
		return errors.New("user must be a manager")
	}

	return s.teamRepo.AddManager(teamID, userID)
}

func (s *TeamService) RemoveManager(teamID, userID, requestorID uuid.UUID) error {
	// Verify requestor has permission
	if err := s.verifyManagerPermission(teamID, requestorID); err != nil {
		return err
	}

	return s.teamRepo.RemoveManager(teamID, userID)
}

func (s *TeamService) GetTeam(teamID uuid.UUID) (*models.Team, error) {
	return s.teamRepo.GetByID(teamID)
}

func (s *TeamService) GetAllTeams() ([]models.Team, error) {
	return s.teamRepo.GetAll()
}

func (s *TeamService) verifyManagerPermission(teamID, userID uuid.UUID) error {
	isManager, err := s.teamRepo.IsManager(teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to check manager status: %w", err)
	}
	if !isManager {
		return errors.New("insufficient permissions: user is not a manager of this team")
	}
	return nil
}
