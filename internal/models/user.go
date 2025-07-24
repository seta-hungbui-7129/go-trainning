package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleManager UserRole = "manager"
	RoleMember  UserRole = "member"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Role         UserRole  `json:"role" gorm:"type:varchar(20);not null;default:'member'"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	OwnedFolders    []Folder    `json:"owned_folders,omitempty" gorm:"foreignKey:OwnerID"`
	OwnedNotes      []Note      `json:"owned_notes,omitempty" gorm:"foreignKey:OwnerID"`
	ManagedTeams    []Team      `json:"managed_teams,omitempty" gorm:"many2many:team_managers;"`
	MemberTeams     []Team      `json:"member_teams,omitempty" gorm:"many2many:team_members;"`
	SharedFolders   []Folder    `json:"shared_folders,omitempty" gorm:"many2many:folder_shares;"`
	SharedNotes     []Note      `json:"shared_notes,omitempty" gorm:"many2many:note_shares;"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (u *User) IsManager() bool {
	return u.Role == RoleManager
}

func (u *User) IsMember() bool {
	return u.Role == RoleMember
}
