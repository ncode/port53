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
			c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), existingBackend.ID))
			return c.String(http.StatusConflict, "Backend already exists")
		} else if status.Error.Error() == "UNIQUE constraint failed: backends.id" {
			c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), backend.ID))
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

func (r *BackendRoute) Update(c echo.Context) (err error) {
	var backend model.Backend
	r.db.First(&backend, "id = ?", c.Param("id"))
	if backend.ID == "" {
		return c.String(http.StatusNotFound, "Backend not found")
	}
	if err := c.Bind(&backend); err != nil {
		return err
	}
	r.db.Model(&backend).Updates(&backend)
	marshal, err := jsonapi.Marshal(backend)
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
	if len(backend.Zones) == 0 {
		backend.Zones = nil
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

func (r *BackendRoute) GetZones(c echo.Context) (err error) {
	var zones []model.Zone
	r.db.Model(&zones).Where("backend_id = ?", c.Param("id")).Preload("Backend").Find(&zones)
	marshal, err := jsonapi.Marshal(zones)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, binder.MIMEApplicationJSONApi, marshal)
}

func (r *BackendRoute) AddZone(c echo.Context) (err error) {
	var backend model.Backend
	r.db.Preload("Zones").First(&backend, "id = ?", c.Param("id"))
	if backend.ID == "" {
		return c.String(http.StatusNotFound, "Backend not found")
	}
	var zone model.Zone
	if err := c.Bind(&zone); err != nil {
		return err
	}
	if zone.ID == "" {
		return c.String(http.StatusBadRequest, "Zone ID is required")
	}
	backend.Zones = append(backend.Zones, &zone)
	err = r.db.Model(&backend).Association("Zones").Append(zone)
	if err != nil {
		return err
	}
	marshal, err := jsonapi.Marshal(backend)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, binder.MIMEApplicationJSONApi, marshal)
}

func (r *BackendRoute) RemoveZone(c echo.Context) (err error) {
	var backend model.Backend
	r.db.Preload("Zones").First(&backend, "id = ?", c.Param("id"))
	if backend.ID == "" {
		return c.String(http.StatusNotFound, "Backend not found")
	}
	var zone model.Zone
	if err := c.Bind(&zone); err != nil {
		return err
	}
	if zone.ID == "" {
		return c.String(http.StatusBadRequest, "Zone ID is required")
	}
	err = r.db.Model(&backend).Association("Zones").Delete(zone)
	if err != nil {
		return err
	}
	marshal, err := jsonapi.Marshal(backend)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, binder.MIMEApplicationJSONApi, marshal)
}

func (r *BackendRoute) UpdateZone(c echo.Context) (err error) {
	var backend model.Backend
	r.db.Preload("Zones").First(&backend, "id = ?", c.Param("id"))
	if backend.ID == "" {
		return c.String(http.StatusNotFound, "Backend not found")
	}
	var zones []model.Zone
	if err := c.Bind(&zones); err != nil {
		return err
	}
	if len(zones) == 0 {
		return c.String(http.StatusBadRequest, "Zone ID is required")
	}
	err = r.db.Model(&backend).Association("Zones").Replace(zones)
	if err != nil {
		return err
	}
	return r.Get(c)
}

func (r *BackendRoute) Register(e *echo.Echo) {
	e.GET("/v1/backends/:id", r.Get)
	e.DELETE("/v1/backends/:id", r.Delete)
	e.POST("/v1/backends", r.Create)
	e.PATCH("/v1/backends/:id", r.Update)
	e.GET("/v1/backends", r.List)
	// Relationships
	e.GET("/v1/backends/:id/zones", r.GetZones)
	e.POST("/v1/backends/:id/zones", r.AddZone)
	e.PATCH("/v1/backends/:id/zones", r.UpdateZone)
	e.DELETE("/v1/backends/:id/zones", r.RemoveZone)

}
