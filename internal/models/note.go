package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Note struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title     string    `json:"title" gorm:"not null"`
	Body      string    `json:"body" gorm:"type:text"`
	FolderID  uuid.UUID `json:"folder_id" gorm:"type:uuid;not null"`
	OwnerID   uuid.UUID `json:"owner_id" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Folder      Folder      `json:"folder,omitempty" gorm:"foreignKey:FolderID"`
	Owner       User        `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	SharedUsers []User      `json:"shared_users,omitempty" gorm:"many2many:note_shares;"`
	Shares      []NoteShare `json:"shares,omitempty" gorm:"foreignKey:NoteID"`
}

func (n *Note) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// NoteShare represents the sharing relationship between notes and users
type NoteShare struct {
	ID        uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	NoteID    uuid.UUID   `json:"note_id" gorm:"type:uuid;not null"`
	UserID    uuid.UUID   `json:"user_id" gorm:"type:uuid;not null"`
	Access    AccessLevel `json:"access" gorm:"type:varchar(10);not null;default:'read'"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`

	// Relationships
	Note Note `json:"note,omitempty" gorm:"foreignKey:NoteID"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (ns *NoteShare) BeforeCreate(tx *gorm.DB) error {
	if ns.ID == uuid.Nil {
		ns.ID = uuid.New()
	}
	return nil
}
