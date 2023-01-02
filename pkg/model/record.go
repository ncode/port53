package model

import "gorm.io/gorm"

type Record struct {
	gorm.Model
	Name string
	TTL  int `gorm:"default:3600"`
	Type string
	Data string
	Zone *Zone `gorm:"foreignKey:ZoneID"`
}
