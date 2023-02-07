package model

import (
	"fmt"
	"time"

	"github.com/DataDog/jsonapi"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Record struct {
	ID        string         `gorm:"primarykey;not null" jsonapi:"primary,records"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created_at,omitempty"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"uniqueIndex;not null" jsonapi:"attribute" json:"name"`
	TTL       int            `gorm:"not null;default:3600" jsonapi:"attribute" json:"ttl"`
	Type      string         `gorm:"not null" jsonapi:"attribute" json:"type"`
	Content   string         `gorm:"not null" jsonapi:"attribute" json:"content"`
	Zone      *Zone          `gorm:"-" jsonapi:"relationship" json:"zones"`
	ZoneID    string         `gorm:"foreignKey:ZoneID" json:"-"`
}

// Link returns the link to the resource
func (r *Record) Link() *jsonapi.Link {
	return &jsonapi.Link{
		Self: fmt.Sprintf("%s/v1/records/%s", viper.GetString("serviceUrl"), r.ID),
	}
}

// LinkRelation returns the link to the related resource
func (r *Record) LinkRelation(relation string) *jsonapi.Link {
	return &jsonapi.Link{
		Self:    fmt.Sprintf("%s/v1/records/%s/relationships/%s", viper.GetString("serviceUrl"), r.ID, relation),
		Related: fmt.Sprintf("%s/v1/records/%s/%s", viper.GetString("serviceUrl"), r.ID, relation),
	}
}

// BeforeCreate generates a new ULID for the record if needed
// TOOO: Implement data validation for the record type and content
func (r *Record) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = ulid.Make().String()
	} else {
		_, err = ulid.Parse(r.ID)
	}
	return err
}

// Get the record
func (r *Record) Get(db *gorm.DB, preload bool) error {
	if preload {
		err := db.First(r, "id = ?", r.ID).Error
		if err != nil {
			return err
		}
		r.Zone = &Zone{}
		return db.First(r.Zone, "id = ?", r.ZoneID).Error
	}
	return db.First(r).Error
}

// Delete the record
func (r *Record) Delete(db *gorm.DB) error {
	return db.Delete(r).Error
}

// Update the record
func (r *Record) Update(db *gorm.DB, record Record) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(r).Updates(record).Error
	})
}

// ReplaceZone replaces the zone of the record
func (r *Record) ReplaceZone(db *gorm.DB, zone *Zone) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(r).Updates(Record{ZoneID: zone.ID}).Error
	})
}

// DeleteZone deletes the zone of the record
func (r *Record) DeleteZone(db *gorm.DB) error {
	return db.Model(r).Association("Zone").Clear()
}
