package model

import (
	"fmt"
	"time"

	"github.com/DataDog/jsonapi"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Zone struct {
	ID        string         `gorm:"primarykey,not null" jsonapi:"primary,zones"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created_at,omitempty"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"uniqueIndex;not null" jsonapi:"attribute" json:"name"`
	TTL       int            `gorm:"default:3600" jsonapi:"attribute" json:"ttl"`
	MName     string         `gorm:"default:@;not null" jsonapi:"attribute" json:"mname"`
	RName     string         `gorm:"default:admin;not null" jsonapi:"attribute" json:"rname"`
	Serial    int            `gorm:"default:1" jsonapi:"attribute" json:"serial"`
	Refresh   int            `gorm:"default:3600" jsonapi:"attribute" json:"refresh"`
	Retry     int            `gorm:"default:600" jsonapi:"attribute" json:"retry"`
	Expire    int            `gorm:"default:604800" jsonapi:"attribute" json:"expire"`
	Minimum   int            `gorm:"default:3600" jsonapi:"attribute" json:"minimum"`
	Records   []*Record      `gorm:"foreignKey:ZoneID" jsonapi:"relationship" json:"records,omitempty"`
	Backends  []*Backend     `gorm:"many2many:backend_zones;" jsonapi:"relationship" json:"backends,omitempty"`
}

// Link returns the link to the resource
func (z *Zone) Link() *jsonapi.Link {
	return &jsonapi.Link{
		Self: fmt.Sprintf("%s/v1/zones/%s", viper.GetString("serviceUrl"), z.ID),
	}
}

// LinkRelation returns the link to the related resource
func (z *Zone) LinkRelation(relation string) *jsonapi.Link {
	return &jsonapi.Link{
		Self:    fmt.Sprintf("%s/v1/zones/%s/relationships/%s", viper.GetString("serviceUrl"), z.ID, relation),
		Related: fmt.Sprintf("%s/v1/zones/%s/%s", viper.GetString("serviceUrl"), z.ID, relation),
	}
}

// BeforeCreate generates a new ULID for the zone if needed
func (z *Zone) BeforeCreate(tx *gorm.DB) (err error) {
	if z.ID == "" {
		z.ID = ulid.Make().String()
	} else {
		_, err = ulid.Parse(z.ID)
	}
	return err
}

// Get a zone
func (z *Zone) Get(db *gorm.DB, preload bool) (err error) {
	if preload {
		return db.Preload("Records").Preload("Backends").First(z, "id = ?", z.ID).Error
	}
	return db.First(z, "id = ?", z.ID).Error
}

// Delete a zone
func (z *Zone) Delete(db *gorm.DB) (err error) {
	return db.Delete(&z).Error
}

// AddBackend adds a zone to the zone
func (z *Zone) AddBackend(db *gorm.DB, backend *Backend) (err error) {
	return db.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(z).Association("Backends").Append(backend)
		if err != nil {
			return err
		}
		return nil
	})
}

// RemoveBackend removes a backend from the zone
func (z *Zone) RemoveBackend(db *gorm.DB, backend *Backend) (err error) {
	return db.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(z).Association("Backends").Delete(backend)
		if err != nil {
			return err
		}
		return nil
	})
}

// ReplaceBackends replaces all backends of the zone
func (z *Zone) ReplaceBackends(db *gorm.DB, backends []*Backend) (err error) {
	return db.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(z).Association("Backends").Replace(backends)
		if err != nil {
			return err
		}
		return nil
	})
}
