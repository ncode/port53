package model

import (
	"fmt"
	"time"

	"github.com/DataDog/jsonapi"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Backend struct {
	ID        string         `gorm:"primarykey'not null" jsonapi:"primary,backends" json:"id,omitempty"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created_at,omitempty"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"uniqueIndex;not null" jsonapi:"attribute" json:"name"`
	Zones     []*Zone        `gorm:"many2many:backend_zones;" jsonapi:"relationship" json:"zones,omitempty"`
}

// Link returns the link to the backend
func (b *Backend) Link() *jsonapi.Link {
	return &jsonapi.Link{
		Self: fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), b.ID),
	}
}

// LinkRelation returns the link to the relationship
func (b *Backend) LinkRelation(relation string) *jsonapi.Link {
	return &jsonapi.Link{
		Self:    fmt.Sprintf("%s/v1/backends/%s/relationships/%s", viper.GetString("serviceUrl"), b.ID, relation),
		Related: fmt.Sprintf("%s/v1/backends/%s/%s", viper.GetString("serviceUrl"), b.ID, relation),
	}
}

// BeforeCreate generates a new ULID for the backend if needed
func (b *Backend) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID = ulid.Make().String()
	} else {
		_, err = ulid.Parse(b.ID)
	}
	return err
}

// Get returns the backend with the given id
func (b *Backend) Get(db *gorm.DB, preload bool) (err error) {
	if preload {
		return db.Preload("Zones").First(b, "id = ?", b.ID).Error
	}
	return db.First(b, "id = ?", b.ID).Error
}

// AddZone adds a zone to the backend
func (b *Backend) AddZone(db *gorm.DB, zone *Zone) (err error) {
	return db.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(b).Association("Zones").Append(zone)
		if err != nil {
			return err
		}
		return nil
	})
}

// RemoveZone removes a zone from the backend
func (b *Backend) RemoveZone(db *gorm.DB, zone *Zone) (err error) {
	return db.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(b).Association("Zones").Delete(zone)
		if err != nil {
			return err
		}
		return nil
	})
}

// ReplaceZones replaces the zones of the backend
func (b *Backend) ReplaceZones(db *gorm.DB, zones []*Zone) (err error) {
	return db.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(b).Association("Zones").Replace(zones)
		if err != nil {
			return err
		}
		return nil
	})
}

// Delete deletes the backend from the database
func (b *Backend) Delete(db *gorm.DB) (err error) {
	return db.Delete(&b).Error
}
