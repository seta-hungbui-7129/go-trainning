package services

import (
	"context"
	"io"
	"github.com/google/uuid"
	"seta-training/internal/models"
	"seta-training/pkg/auth"
)

// UserServiceInterface defines the interface for user service
type UserServiceInterface interface {
	CreateUser(input *CreateUserInput) (*models.User, error)
	Login(input *LoginInput) (*LoginResponse, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	ValidateToken(tokenString string) (*auth.Claims, error)
}

// TeamServiceInterface defines the interface for team service
type TeamServiceInterface interface {
	CreateTeam(input *CreateTeamInput, creatorID uuid.UUID) (*models.Team, error)
	AddMember(teamID, userID, managerID uuid.UUID) error
	RemoveMember(teamID, userID, managerID uuid.UUID) error
	AddManager(teamID, userID, requestorID uuid.UUID) error
	RemoveManager(teamID, userID, requestorID uuid.UUID) error
	GetTeam(teamID uuid.UUID) (*models.Team, error)
	GetAllTeams() ([]models.Team, error)
}

// FolderServiceInterface defines the interface for folder service
type FolderServiceInterface interface {
	CreateFolder(input *CreateFolderInput, ownerID uuid.UUID) (*models.Folder, error)
	GetFolder(folderID, userID uuid.UUID) (*models.Folder, error)
	UpdateFolder(folderID uuid.UUID, input *UpdateFolderInput, userID uuid.UUID) (*models.Folder, error)
	DeleteFolder(folderID, userID uuid.UUID) error
	ShareFolder(folderID uuid.UUID, input *ShareFolderInput, ownerID uuid.UUID) error
	RevokeShare(folderID, targetUserID, ownerID uuid.UUID) error
	GetUserFolders(userID uuid.UUID) ([]models.Folder, error)
}

// NoteServiceInterface defines the interface for note service
type NoteServiceInterface interface {
	CreateNote(folderID uuid.UUID, input *CreateNoteInput, userID uuid.UUID) (*models.Note, error)
	GetNote(noteID, userID uuid.UUID) (*models.Note, error)
	UpdateNote(noteID uuid.UUID, input *UpdateNoteInput, userID uuid.UUID) (*models.Note, error)
	DeleteNote(noteID, userID uuid.UUID) error
	ShareNote(noteID uuid.UUID, input *ShareNoteInput, ownerID uuid.UUID) error
	RevokeShare(noteID, targetUserID, ownerID uuid.UUID) error
	GetUserNotes(userID uuid.UUID) ([]models.Note, error)
}

// ImportServiceInterface defines the interface for import service
type ImportServiceInterface interface {
	ImportUsersFromCSV(ctx context.Context, csvReader io.Reader, config ImportConfig) (*ImportSummary, error)
}
