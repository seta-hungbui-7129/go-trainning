package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"seta-training/internal/models"
)

type NoteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) Create(note *models.Note) error {
	return r.db.Create(note).Error
}

func (r *NoteRepository) GetByID(id uuid.UUID) (*models.Note, error) {
	var note models.Note
	err := r.db.Preload("Owner").Preload("Folder").Preload("Shares.User").Where("id = ?", id).First(&note).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("note not found")
		}
		return nil, err
	}
	return &note, nil
}

func (r *NoteRepository) GetByFolder(folderID uuid.UUID) ([]models.Note, error) {
	var notes []models.Note
	err := r.db.Where("folder_id = ?", folderID).Preload("Owner").Find(&notes).Error
	return notes, err
}

func (r *NoteRepository) GetByOwner(ownerID uuid.UUID) ([]models.Note, error) {
	var notes []models.Note
	err := r.db.Where("owner_id = ?", ownerID).Preload("Folder").Find(&notes).Error
	return notes, err
}

func (r *NoteRepository) Update(note *models.Note) error {
	return r.db.Save(note).Error
}

func (r *NoteRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Note{}, id).Error
}

func (r *NoteRepository) ShareNote(noteID, userID uuid.UUID, access models.AccessLevel) error {
	share := &models.NoteShare{
		NoteID: noteID,
		UserID: userID,
		Access: access,
	}
	return r.db.Create(share).Error
}

func (r *NoteRepository) RevokeShare(noteID, userID uuid.UUID) error {
	return r.db.Where("note_id = ? AND user_id = ?", noteID, userID).Delete(&models.NoteShare{}).Error
}

func (r *NoteRepository) GetSharedNotes(userID uuid.UUID) ([]models.Note, error) {
	var notes []models.Note
	err := r.db.Joins("JOIN note_shares ON notes.id = note_shares.note_id").
		Where("note_shares.user_id = ?", userID).
		Preload("Owner").Preload("Folder").Preload("Shares.User").
		Find(&notes).Error
	return notes, err
}

func (r *NoteRepository) GetUserAccess(noteID, userID uuid.UUID) (*models.NoteShare, error) {
	var share models.NoteShare
	err := r.db.Where("note_id = ? AND user_id = ?", noteID, userID).First(&share).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &share, nil
}

func (r *NoteRepository) HasAccess(noteID, userID uuid.UUID) (bool, models.AccessLevel, error) {
	// Check if user is owner
	var note models.Note
	err := r.db.Where("id = ? AND owner_id = ?", noteID, userID).First(&note).Error
	if err == nil {
		return true, models.AccessWrite, nil
	}

	// Check if user has shared access
	share, err := r.GetUserAccess(noteID, userID)
	if err != nil {
		return false, "", err
	}
	if share != nil {
		return true, share.Access, nil
	}

	return false, "", nil
}
