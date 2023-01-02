package api

import (
	"github.com/DataDog/jsonapi"
	"net/http"

	"github.com/ncode/trutinha/pkg/model"
	"github.com/oklog/ulid/v2"

	"github.com/labstack/echo/v4"
)

type BackendRoute struct {
	//db *gorm.DB
}

func (r *BackendRoute) Create(c echo.Context) (err error) {
	var backend model.Backend
	if err := c.Bind(&backend); err != nil {
		return err
	}
	if backend.ID == "" {
		backend.ID = ulid.Make().String()
	}
	//r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&backend)
	marshal, err := jsonapi.Marshal(backend)
	if err != nil {
		return err
	}
	return c.JSONBlob(http.StatusCreated, marshal)
}

func (r *BackendRoute) Register(e *echo.Echo) {
	e.POST("/v1/backend", r.Create)
}
