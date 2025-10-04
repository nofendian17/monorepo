package model

import (
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Agent struct {
	ID            string         `gorm:"type:char(26);primaryKey"`
	AgentName     string         `gorm:"type:varchar(255);not null"`
	AgentType     string         `gorm:"type:varchar(20);not null;check:agent_type IN ('IATA','SUB_AGENT')"`
	ParentAgentID *string        `gorm:"type:char(26);default:null"`
	Parent        *Agent         `gorm:"foreignKey:ParentAgentID;references:ID"`
	Children      []Agent        `gorm:"foreignKey:ParentAgentID"`
	Email         string         `gorm:"type:varchar(255);not null"`
	IsActive      bool           `gorm:"default:true"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (a *Agent) BeforeCreate(tx *gorm.DB) error {
	a.ID = ulid.Make().String()
	return nil
}
