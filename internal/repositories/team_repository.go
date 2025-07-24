package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"seta-training/internal/models"
)

type TeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(team *models.Team) error {
	return r.db.Create(team).Error
}

func (r *TeamRepository) GetByID(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.db.Preload("Managers").Preload("Members").Where("id = ?", id).First(&team).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("team not found")
		}
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) GetAll() ([]models.Team, error) {
	var teams []models.Team
	err := r.db.Preload("Managers").Preload("Members").Find(&teams).Error
	return teams, err
}

func (r *TeamRepository) Update(team *models.Team) error {
	return r.db.Save(team).Error
}

func (r *TeamRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Team{}, id).Error
}

func (r *TeamRepository) AddManager(teamID, userID uuid.UUID) error {
	return r.db.Create(&models.TeamManager{
		TeamID: teamID,
		UserID: userID,
	}).Error
}

func (r *TeamRepository) RemoveManager(teamID, userID uuid.UUID) error {
	return r.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&models.TeamManager{}).Error
}

func (r *TeamRepository) AddMember(teamID, userID uuid.UUID) error {
	return r.db.Create(&models.TeamMember{
		TeamID: teamID,
		UserID: userID,
	}).Error
}

func (r *TeamRepository) RemoveMember(teamID, userID uuid.UUID) error {
	return r.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&models.TeamMember{}).Error
}

func (r *TeamRepository) IsManager(teamID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.TeamManager{}).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count).Error
	return count > 0, err
}

func (r *TeamRepository) IsMember(teamID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.TeamMember{}).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count).Error
	return count > 0, err
}

func (r *TeamRepository) GetTeamsByManager(userID uuid.UUID) ([]models.Team, error) {
	var teams []models.Team
	err := r.db.Joins("JOIN team_managers ON teams.id = team_managers.team_id").
		Where("team_managers.user_id = ?", userID).
		Preload("Managers").Preload("Members").
		Find(&teams).Error
	return teams, err
}

func (r *TeamRepository) GetTeamsByMember(userID uuid.UUID) ([]models.Team, error) {
	var teams []models.Team
	err := r.db.Joins("JOIN team_members ON teams.id = team_members.team_id").
		Where("team_members.user_id = ?", userID).
		Preload("Managers").Preload("Members").
		Find(&teams).Error
	return teams, err
}
