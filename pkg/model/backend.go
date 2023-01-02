package model

import "gorm.io/gorm"

type Backend struct {
	gorm.Model
	Name  string  `gorm:"uniqueIndex"`
	Zones []*Zone `gorm:"many2many:backend_zones;"`
}
