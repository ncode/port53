package model

import (
	"gorm.io/gorm"
	"time"
)

type Backend struct {
	ID        string         `gorm:"primarykey" jsonapi:"primary,backend" json:"id,omitempty"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created_at"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"uniqueIndex" jsonapi:"attribute" json:"name"`
	Zones     []*Zone        `gorm:"many2many:backend_zones;" jsonapi:"relationship" json:"zones,omitempty"`
}
