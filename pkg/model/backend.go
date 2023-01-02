package model

import (
	"gorm.io/gorm"
	"time"
)

type Backend struct {
	ID        uint           `gorm:"primarykey" jsonapi:"primary,backend"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created_at"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" jsonapi:"attribute" json:"deleted_at"`
	Name      string         `gorm:"uniqueIndex" jsonapi:"attribute" json:"name"`
	Zones     []*Zone        `gorm:"many2many:backend_zones;" jsonapi:"relationship" json:"zones"`
}
