package api

import (
	"fmt"
	"github.com/DataDog/jsonapi"
	"github.com/ncode/trutinha/pkg/binder"
	"gorm.io/gorm"
	"net/http"

	"github.com/ncode/trutinha/pkg/model"
	"github.com/oklog/ulid/v2"

	"github.com/labstack/echo/v4"
)

type BackendRoute struct {
	db *gorm.DB
}

func (r *BackendRoute) Create(c echo.Context) (err error) {
	var backend model.Backend
	if err := c.Bind(&backend); err != nil {
		return err
	}
	if backend.Name == "" {
		return c.String(http.StatusBadRequest, "Name is required")
	}
	if backend.ID == "" {
		backend.ID = ulid.Make().String()
	}
	status := r.db.Create(&backend)
	if status.Error != nil {
		if status.Error.Error() == "UNIQUE constraint failed: backends.name" {
			var existingBackend model.Backend
			r.db.First(&existingBackend, "name = ?", backend.Name)
			c.Response().Header().Set("Location", fmt.Sprintf("/v1/backend/%s", existingBackend.ID))
			return c.String(http.StatusConflict, "Backend already exists")
		} else if status.Error.Error() == "UNIQUE constraint failed: backends.id" {
			c.Response().Header().Set("Location", fmt.Sprintf("/v1/backend/%s", backend.ID))
			return c.String(http.StatusConflict, "Backend already exists")
		}
		return c.String(http.StatusInternalServerError, status.Error.Error())
	}
	marshal, err := jsonapi.Marshal(backend)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusCreated, binder.MIMEApplicationJSONApi, marshal)
}

func (r *BackendRoute) Get(c echo.Context) (err error) {
	var backend model.Backend
	r.db.Preload("Zones").First(&backend, "id = ?", c.Param("id"))
	if backend.ID == "" {
		return c.String(http.StatusNotFound, "Not found")
	}
	marshal, err := jsonapi.Marshal(backend)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, binder.MIMEApplicationJSONApi, marshal)
}

func (r *BackendRoute) Delete(c echo.Context) (err error) {
	r.db.Where("id = ?", c.Param("id")).Delete(&model.Backend{})
	return c.NoContent(http.StatusNoContent)
}

func (r *BackendRoute) Register(e *echo.Echo) {
	e.GET("/v1/backend/:id", r.Get)
	e.POST("/v1/backend", r.Create)
}
