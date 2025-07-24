package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"seta-training/internal/models"
)

type FolderRepository struct {
	db *gorm.DB
}

func NewFolderRepository(db *gorm.DB) *FolderRepository {
	return &FolderRepository{db: db}
}

func (r *FolderRepository) Create(folder *models.Folder) error {
	return r.db.Create(folder).Error
}

func (r *FolderRepository) GetByID(id uuid.UUID) (*models.Folder, error) {
	var folder models.Folder
	err := r.db.Preload("Owner").Preload("Notes").Preload("Shares.User").Where("id = ?", id).First(&folder).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("folder not found")
		}
		return nil, err
	}
	return &folder, nil
}

func (r *FolderRepository) GetByOwner(ownerID uuid.UUID) ([]models.Folder, error) {
	var folders []models.Folder
	err := r.db.Where("owner_id = ?", ownerID).Preload("Notes").Find(&folders).Error
	return folders, err
}

func (r *FolderRepository) Update(folder *models.Folder) error {
	return r.db.Save(folder).Error
}

func (r *FolderRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Folder{}, id).Error
}

func (r *FolderRepository) ShareFolder(folderID, userID uuid.UUID, access models.AccessLevel) error {
	share := &models.FolderShare{
		FolderID: folderID,
		UserID:   userID,
		Access:   access,
	}
	return r.db.Create(share).Error
}

func (r *FolderRepository) RevokeShare(folderID, userID uuid.UUID) error {
	return r.db.Where("folder_id = ? AND user_id = ?", folderID, userID).Delete(&models.FolderShare{}).Error
}

func (r *FolderRepository) GetSharedFolders(userID uuid.UUID) ([]models.Folder, error) {
	var folders []models.Folder
	err := r.db.Joins("JOIN folder_shares ON folders.id = folder_shares.folder_id").
		Where("folder_shares.user_id = ?", userID).
		Preload("Owner").Preload("Notes").Preload("Shares.User").
		Find(&folders).Error
	return folders, err
}

func (r *FolderRepository) GetUserAccess(folderID, userID uuid.UUID) (*models.FolderShare, error) {
	var share models.FolderShare
	err := r.db.Where("folder_id = ? AND user_id = ?", folderID, userID).First(&share).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &share, nil
}

func (r *FolderRepository) HasAccess(folderID, userID uuid.UUID) (bool, models.AccessLevel, error) {
	// Check if user is owner
	var folder models.Folder
	err := r.db.Where("id = ? AND owner_id = ?", folderID, userID).First(&folder).Error
	if err == nil {
		return true, models.AccessWrite, nil
	}

	// Check if user has shared access
	share, err := r.GetUserAccess(folderID, userID)
	if err != nil {
		return false, "", err
	}
	if share != nil {
		return true, share.Access, nil
	}

	return false, "", nil
}
