// Package model contains data models for the application
package model

import (
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

// User represents a user in the system
// It contains the essential fields that define a user entity
type User struct {
	// ID is the unique identifier for the user
	ID string `gorm:"type:char(26);primaryKey"`
	// AgentID is the identifier of the agent associated with the user
	AgentID *string `gorm:"type:char(26);index"`
	// Agent represents the associated agent entity
	Agent Agent `gorm:"references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	// Name is the user's full name
	Name string `gorm:"not null"`
	// Email is the user's email address which must be unique
	Email string `gorm:"uniqueIndex;not null"`
	// IsActive indicates whether the user is active
	IsActive bool `gorm:"default:true"`
	// CreatedAt is the timestamp when the user was created
	CreatedAt time.Time `gorm:"autoCreateTime"`
	// UpdatedAt is the timestamp when the user was last updated
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	// DeletedAt is the timestamp when the user was soft deleted (for soft delete functionality)
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = ulid.Make().String()
	return nil
}
