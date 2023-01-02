package model

import (
	"gorm.io/gorm"
	"time"
)

type Zone struct {
	ID        uint           `gorm:"primarykey" jsonapi:"primary,record"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created_at"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" jsonapi:"attribute" json:"deleted_at"`
	Name      string         `gorm:"uniqueIndex" jsonapi:"attribute" json:"name"`
	TTL       int            `gorm:"default:3600" jsonapi:"attribute" json:"ttl"`
	MName     string         `gorm:"default:@" jsonapi:"attribute" json:"mname"`
	RName     string         `gorm:"default:admin" jsonapi:"attribute" json:"rname"`
	Serial    int            `gorm:"default:1" jsonapi:"attribute" json:"serial"`
	Refresh   int            `gorm:"default:3600" jsonapi:"attribute" json:"refresh"`
	Retry     int            `gorm:"default:600" jsonapi:"attribute" json:"retry"`
	Expire    int            `gorm:"default:604800" jsonapi:"attribute" json:"expire"`
	Minimum   int            `gorm:"default:3600" jsonapi:"attribute" json:"minimum"`
	Records   []*Record      `gorm:"foreignKey:ZoneID" jsonapi:"relationship" json:"records"`
	Backends  []*Backend     `gorm:"many2many:backend_zones;" jsonapi:"relationship" json:"backends"`
}
