package model

import (
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

// Supplier represents a supplier in the system
type Supplier struct {
	ID           int            `gorm:"primaryKey;autoIncrement"`
	SupplierCode string         `gorm:"type:varchar(50);unique;not null"`
	SupplierName string         `gorm:"type:varchar(100);not null"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// AgentSupplierCredential represents the credentials for an agent-supplier pair
type AgentSupplierCredential struct {
	ID          string         `gorm:"type:char(26);primaryKey"`
	IataAgentID string         `gorm:"type:char(26);not null;uniqueIndex:iata_agent_id_supplier_id"`
	SupplierID  int            `gorm:"not null;uniqueIndex:iata_agent_id_supplier_id"`
	Supplier    Supplier       `gorm:"foreignKey:SupplierID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Credentials string         `gorm:"type:text;not null"` // Encrypted JSON
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (a *AgentSupplierCredential) BeforeCreate(tx *gorm.DB) error {
	a.ID = ulid.Make().String()
	return nil
}
