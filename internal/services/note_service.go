package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"seta-training/internal/models"
	"seta-training/internal/repositories"
)

type NoteService struct {
	noteRepo   repositories.NoteRepositoryInterface
	folderRepo repositories.FolderRepositoryInterface
}

func NewNoteService(noteRepo repositories.NoteRepositoryInterface, folderRepo repositories.FolderRepositoryInterface) *NoteService {
	return &NoteService{
		noteRepo:   noteRepo,
		folderRepo: folderRepo,
	}
}

type CreateNoteInput struct {
	Title string `json:"title" binding:"required,min=1,max=200"`
	Body  string `json:"body"`
}

type UpdateNoteInput struct {
	Title string `json:"title" binding:"required,min=1,max=200"`
	Body  string `json:"body"`
}

type ShareNoteInput struct {
	UserID uuid.UUID          `json:"userId" binding:"required"`
	Access models.AccessLevel `json:"access" binding:"required,oneof=read write"`
}

func (s *NoteService) CreateNote(folderID uuid.UUID, input *CreateNoteInput, userID uuid.UUID) (*models.Note, error) {
	// Check if user has write access to the folder
	hasAccess, access, err := s.folderRepo.HasAccess(folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check folder access: %w", err)
	}
	if !hasAccess || access != models.AccessWrite {
		return nil, errors.New("write access to folder required")
	}

	note := &models.Note{
		Title:    input.Title,
		Body:     input.Body,
		FolderID: folderID,
		OwnerID:  userID,
	}

	if err := s.noteRepo.Create(note); err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return s.noteRepo.GetByID(note.ID)
}

func (s *NoteService) GetNote(noteID, userID uuid.UUID) (*models.Note, error) {
	// Check if user has access to the note
	hasAccess, _, err := s.noteRepo.HasAccess(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess {
		return nil, errors.New("access denied")
	}

	return s.noteRepo.GetByID(noteID)
}

func (s *NoteService) UpdateNote(noteID uuid.UUID, input *UpdateNoteInput, userID uuid.UUID) (*models.Note, error) {
	// Check if user has write access
	hasAccess, access, err := s.noteRepo.HasAccess(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess || access != models.AccessWrite {
		return nil, errors.New("write access required")
	}

	note, err := s.noteRepo.GetByID(noteID)
	if err != nil {
		return nil, err
	}

	note.Title = input.Title
	note.Body = input.Body
	if err := s.noteRepo.Update(note); err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return note, nil
}

func (s *NoteService) DeleteNote(noteID, userID uuid.UUID) error {
	// Only owner can delete note
	note, err := s.noteRepo.GetByID(noteID)
	if err != nil {
		return err
	}
	if note.OwnerID != userID {
		return errors.New("only owner can delete note")
	}

	return s.noteRepo.Delete(noteID)
}

func (s *NoteService) ShareNote(noteID uuid.UUID, input *ShareNoteInput, ownerID uuid.UUID) error {
	// Only owner can share note
	note, err := s.noteRepo.GetByID(noteID)
	if err != nil {
		return err
	}
	if note.OwnerID != ownerID {
		return errors.New("only owner can share note")
	}

	return s.noteRepo.ShareNote(noteID, input.UserID, input.Access)
}

func (s *NoteService) RevokeShare(noteID, targetUserID, ownerID uuid.UUID) error {
	// Only owner can revoke sharing
	note, err := s.noteRepo.GetByID(noteID)
	if err != nil {
		return err
	}
	if note.OwnerID != ownerID {
		return errors.New("only owner can revoke sharing")
	}

	return s.noteRepo.RevokeShare(noteID, targetUserID)
}

func (s *NoteService) GetUserNotes(userID uuid.UUID) ([]models.Note, error) {
	// Get owned notes
	ownedNotes, err := s.noteRepo.GetByOwner(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owned notes: %w", err)
	}

	// Get shared notes
	sharedNotes, err := s.noteRepo.GetSharedNotes(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared notes: %w", err)
	}

	// Combine and return
	allNotes := append(ownedNotes, sharedNotes...)
	return allNotes, nil
}
