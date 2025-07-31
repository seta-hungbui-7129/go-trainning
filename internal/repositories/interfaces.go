package repositories

import (
	"github.com/google/uuid"
	"seta-training/internal/models"
)

// UserRepositoryInterface defines the interface for user repository
type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll() ([]models.User, error)
	EmailExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
}

// TeamRepositoryInterface defines the interface for team repository
type TeamRepositoryInterface interface {
	Create(team *models.Team) error
	GetByID(id uuid.UUID) (*models.Team, error)
	GetAll() ([]models.Team, error)
	AddManager(teamID, userID uuid.UUID) error
	RemoveManager(teamID, userID uuid.UUID) error
	AddMember(teamID, userID uuid.UUID) error
	RemoveMember(teamID, userID uuid.UUID) error
	IsManager(teamID, userID uuid.UUID) (bool, error)
}

// FolderRepositoryInterface defines the interface for folder repository
type FolderRepositoryInterface interface {
	Create(folder *models.Folder) error
	GetByID(id uuid.UUID) (*models.Folder, error)
	GetByOwner(ownerID uuid.UUID) ([]models.Folder, error)
	Update(folder *models.Folder) error
	Delete(id uuid.UUID) error
	ShareFolder(folderID, userID uuid.UUID, access models.AccessLevel) error
	RevokeShare(folderID, userID uuid.UUID) error
	HasAccess(folderID, userID uuid.UUID) (bool, models.AccessLevel, error)
	GetSharedFolders(userID uuid.UUID) ([]models.Folder, error)
}

// NoteRepositoryInterface defines the interface for note repository
type NoteRepositoryInterface interface {
	Create(note *models.Note) error
	GetByID(id uuid.UUID) (*models.Note, error)
	GetByOwner(ownerID uuid.UUID) ([]models.Note, error)
	GetByFolder(folderID uuid.UUID) ([]models.Note, error)
	Update(note *models.Note) error
	Delete(id uuid.UUID) error
	ShareNote(noteID, userID uuid.UUID, access models.AccessLevel) error
	RevokeShare(noteID, userID uuid.UUID) error
	HasAccess(noteID, userID uuid.UUID) (bool, models.AccessLevel, error)
	GetSharedNotes(userID uuid.UUID) ([]models.Note, error)
}
