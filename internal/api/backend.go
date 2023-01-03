package api

import (
	"fmt"
	"github.com/DataDog/jsonapi"
	"github.com/ncode/trutinha/pkg/binder"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ncode/trutinha/pkg/model"
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
	status := r.db.Create(&backend)
	if status.Error != nil {
		if status.Error.Error() == "UNIQUE constraint failed: backends.name" {
			var existingBackend model.Backend
			r.db.First(&existingBackend, "name = ?", backend.Name)
			c.Response().Header().Set("Location", fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), existingBackend.ID))
			return c.String(http.StatusConflict, "Backend already exists")
		} else if status.Error.Error() == "UNIQUE constraint failed: backends.id" {
			c.Response().Header().Set("Location", fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), backend.ID))
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

func (r *BackendRoute) List(c echo.Context) (err error) {
	var backends []model.Backend
	r.db.Preload("Zones").Find(&backends)
	marshal, err := jsonapi.Marshal(backends)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, binder.MIMEApplicationJSONApi, marshal)
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
	e.GET("/v1/backends/:id", r.Get)
	e.DELETE("/v1/backends/:id", r.Delete)
	e.POST("/v1/backends", r.Create)
	e.GET("/v1/backends", r.List)
}
