package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"seta-training/internal/models"
	"seta-training/internal/repositories"
)

type FolderService struct {
	folderRepo *repositories.FolderRepository
	noteRepo   *repositories.NoteRepository
}

func NewFolderService(folderRepo *repositories.FolderRepository, noteRepo *repositories.NoteRepository) *FolderService {
	return &FolderService{
		folderRepo: folderRepo,
		noteRepo:   noteRepo,
	}
}

type CreateFolderInput struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type UpdateFolderInput struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type ShareFolderInput struct {
	UserID uuid.UUID           `json:"userId" binding:"required"`
	Access models.AccessLevel  `json:"access" binding:"required,oneof=read write"`
}

func (s *FolderService) CreateFolder(input *CreateFolderInput, ownerID uuid.UUID) (*models.Folder, error) {
	folder := &models.Folder{
		Name:    input.Name,
		OwnerID: ownerID,
	}

	if err := s.folderRepo.Create(folder); err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	return s.folderRepo.GetByID(folder.ID)
}

func (s *FolderService) GetFolder(folderID, userID uuid.UUID) (*models.Folder, error) {
	// Check if user has access to the folder
	hasAccess, _, err := s.folderRepo.HasAccess(folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess {
		return nil, errors.New("access denied")
	}

	return s.folderRepo.GetByID(folderID)
}

func (s *FolderService) UpdateFolder(folderID uuid.UUID, input *UpdateFolderInput, userID uuid.UUID) (*models.Folder, error) {
	// Check if user has write access
	hasAccess, access, err := s.folderRepo.HasAccess(folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess || access != models.AccessWrite {
		return nil, errors.New("write access required")
	}

	folder, err := s.folderRepo.GetByID(folderID)
	if err != nil {
		return nil, err
	}

	folder.Name = input.Name
	if err := s.folderRepo.Update(folder); err != nil {
		return nil, fmt.Errorf("failed to update folder: %w", err)
	}

	return folder, nil
}

func (s *FolderService) DeleteFolder(folderID, userID uuid.UUID) error {
	// Only owner can delete folder
	folder, err := s.folderRepo.GetByID(folderID)
	if err != nil {
		return err
	}
	if folder.OwnerID != userID {
		return errors.New("only owner can delete folder")
	}

	// Delete all notes in the folder first
	notes, err := s.noteRepo.GetByFolder(folderID)
	if err != nil {
		return fmt.Errorf("failed to get notes: %w", err)
	}

	for _, note := range notes {
		if err := s.noteRepo.Delete(note.ID); err != nil {
			return fmt.Errorf("failed to delete note: %w", err)
		}
	}

	return s.folderRepo.Delete(folderID)
}

func (s *FolderService) ShareFolder(folderID uuid.UUID, input *ShareFolderInput, ownerID uuid.UUID) error {
	// Only owner can share folder
	folder, err := s.folderRepo.GetByID(folderID)
	if err != nil {
		return err
	}
	if folder.OwnerID != ownerID {
		return errors.New("only owner can share folder")
	}

	return s.folderRepo.ShareFolder(folderID, input.UserID, input.Access)
}

func (s *FolderService) RevokeShare(folderID, targetUserID, ownerID uuid.UUID) error {
	// Only owner can revoke sharing
	folder, err := s.folderRepo.GetByID(folderID)
	if err != nil {
		return err
	}
	if folder.OwnerID != ownerID {
		return errors.New("only owner can revoke sharing")
	}

	return s.folderRepo.RevokeShare(folderID, targetUserID)
}

func (s *FolderService) GetUserFolders(userID uuid.UUID) ([]models.Folder, error) {
	// Get owned folders
	ownedFolders, err := s.folderRepo.GetByOwner(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owned folders: %w", err)
	}

	// Get shared folders
	sharedFolders, err := s.folderRepo.GetSharedFolders(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared folders: %w", err)
	}

	// Combine and return
	allFolders := append(ownedFolders, sharedFolders...)
	return allFolders, nil
}
