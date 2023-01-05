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
	CreatedAt time.Time      `jsonapi:"attribute" json:"created_at"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"uniqueIndex;not null" jsonapi:"attribute" json:"name"`
	Zones     []*Zone        `gorm:"many2many:backend_zones;" jsonapi:"relationship" json:"zones,omitempty"`
}

func (b *Backend) Link() *jsonapi.Link {
	return &jsonapi.Link{
		Self: fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), b.ID),
		Related: &jsonapi.LinkObject{
			Href: fmt.Sprintf("%s/v1/backend/%s/zones", viper.GetString("serviceUrl"), b.ID),
			Meta: map[string]int{
				"count": len(b.Zones),
			},
		},
	}
}

func (b *Backend) LinkRelation(relation string) *jsonapi.Link {
	return &jsonapi.Link{
		Self:    fmt.Sprintf("%s/v1/backends/%s/relationships/%s", viper.GetString("serviceUrl"), b.ID, relation),
		Related: fmt.Sprintf("%s/v1/backends/%s/%s", viper.GetString("serviceUrl"), b.ID, relation),
	}
}

func (b *Backend) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID = ulid.Make().String()
	} else {
		_, err = ulid.Parse(b.ID)
	}
	return err
}
