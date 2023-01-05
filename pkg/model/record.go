package model

import (
	"fmt"
	"github.com/DataDog/jsonapi"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"time"
)

type Record struct {
	ID        string         `gorm:"primarykey;not null" jsonapi:"primary,records"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created_at"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"uniqueIndex;not null" jsonapi:"attribute" json:"name"`
	TTL       int            `gorm:"default:3600" jsonapi:"attribute" json:"ttl"`
	Type      string         `gorm:"not null" jsonapi:"attribute" json:"type"`
	Data      string         `gorm:"not null" jsonapi:"attribute" json:"data"`
	ZoneID    string         `gorm:"foreignKey:ZoneID" jsonapi:"relationship" json:"zone,omitempty"`
}

func (r *Record) LinkRelation(relation string) *jsonapi.Link {
	return &jsonapi.Link{
		Self:    fmt.Sprintf("%s/v1/records/%s/relationships/%s", viper.GetString("serviceUrl"), r.ID, relation),
		Related: fmt.Sprintf("%s/v1/records/%s/%s", viper.GetString("serviceUrl"), r.ID, relation),
	}
}

func (r *Record) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = ulid.Make().String()
	} else {
		_, err = ulid.Parse(r.ID)
	}
	return err
}
