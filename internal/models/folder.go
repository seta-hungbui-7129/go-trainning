package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccessLevel string

const (
	AccessRead  AccessLevel = "read"
	AccessWrite AccessLevel = "write"
)

type Folder struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null"`
	OwnerID   uuid.UUID `json:"owner_id" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Owner       User         `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Notes       []Note       `json:"notes,omitempty" gorm:"foreignKey:FolderID"`
	SharedUsers []User       `json:"shared_users,omitempty" gorm:"many2many:folder_shares;"`
	Shares      []FolderShare `json:"shares,omitempty" gorm:"foreignKey:FolderID"`
}

func (f *Folder) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// FolderShare represents the sharing relationship between folders and users
type FolderShare struct {
	ID        uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FolderID  uuid.UUID   `json:"folder_id" gorm:"type:uuid;not null"`
	UserID    uuid.UUID   `json:"user_id" gorm:"type:uuid;not null"`
	Access    AccessLevel `json:"access" gorm:"type:varchar(10);not null;default:'read'"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`

	// Relationships
	Folder Folder `json:"folder,omitempty" gorm:"foreignKey:FolderID"`
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (fs *FolderShare) BeforeCreate(tx *gorm.DB) error {
	if fs.ID == uuid.Nil {
		fs.ID = uuid.New()
	}
	return nil
}
