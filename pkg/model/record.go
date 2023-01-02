package model

import (
	"gorm.io/gorm"
	"time"
)

type Record struct {
	ID        string         `gorm:"primarykey" jsonapi:"primary,record"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created_at"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" jsonapi:"attribute" json:"deleted_at"`
	Name      string         `gorm:"uniqueIndex" jsonapi:"attribute" json:"name"`
	TTL       int            `gorm:"default:3600" jsonapi:"attribute" json:"ttl"`
	Type      string         `jsonapi:"attribute" json:"type"`
	Data      string         `jsonapi:"attribute" json:"data"`
	Zone      *Zone          `gorm:"foreignKey:ZoneID" jsonapi:"relationship" json:"zone"`
}
