package model

import "gorm.io/gorm"

type Zone struct {
	gorm.Model
	Name     string     `gorm:"uniqueIndex"`
	TTL      int        `gorm:"default:3600"`
	MName    string     `gorm:"default:@"`
	RName    string     `gorm:"default:admin"`
	Serial   int        `gorm:"default:1"`
	Refresh  int        `gorm:"default:3600"`
	Retry    int        `gorm:"default:600"`
	Expire   int        `gorm:"default:604800"`
	Minimum  int        `gorm:"default:3600"`
	Records  []*Record  `gorm:"foreignKey:ZoneID"`
	Backends []*Backend `gorm:"many2many:backend_zones;"`
}
