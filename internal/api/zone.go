package api

import (
	"fmt"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/ncode/trutinha/pkg/binder"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
	"github.com/ncode/trutinha/pkg/model"
)

type ZoneRoute struct {
	db *gorm.DB
}

func (r *ZoneRoute) Create(c echo.Context) (err error) {
	var zone model.Zone
	if err := c.Bind(&zone); err != nil {
		return err
	}
	if zone.Name == "" {
		return c.String(http.StatusBadRequest, "Name is required")
	}
	err = r.db.Create(&zone).Error
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: zones.name" {
			var existingZone model.Zone
			err = r.db.First(&existingZone, "name = ?", zone.Name).Error
			if err != nil {
				return err
			}
			c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("%s/v1/zones/%s", viper.GetString("serviceUrl"), existingZone.ID))
			return c.String(http.StatusConflict, "Zone already exists")
		} else if err.Error() == "UNIQUE constraint failed: zones.id" {
			c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("%s/v1/zones/%s", viper.GetString("serviceUrl"), zone.ID))
			return c.String(http.StatusConflict, "Zone already exists")
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}
	marshal, err := jsonapi.Marshal(zone)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusCreated, binder.MIMEApplicationJSONApi, marshal)
}

func (r *ZoneRoute) List(c echo.Context) (err error) {
	var zones []model.Zone
	err = r.db.Preload("Backends").Find(&zones).Error
	if err != nil {
		return err
	}
	for pos, zone := range zones {
		if len(zone.Backends) == 0 {
			zones[pos].Backends = nil
		}
	}
	marshal, err := jsonapi.Marshal(zones)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, binder.MIMEApplicationJSONApi, marshal)
}

func (r *ZoneRoute) Update(c echo.Context) (err error) {
	var zone model.Zone
	err = r.db.First(&zone, "id = ?", c.Param("id")).Error
	if err != nil {
		return err
	}
	if zone.ID == "" {
		return c.String(http.StatusNotFound, "Zone not found")
	}
	if err := c.Bind(&zone); err != nil {
		return err
	}
	r.db.Model(&zone).Updates(&zone)
	marshal, err := jsonapi.Marshal(zone)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, binder.MIMEApplicationJSONApi, marshal)
}

func (r *ZoneRoute) Get(c echo.Context) (err error) {
	var zone model.Zone
	err = r.db.Preload("Backends").First(&zone, "id = ?", c.Param("id")).Error
	if err != nil {
		return err
	}
	if zone.ID == "" {
		return c.String(http.StatusNotFound, "Not found")
	}
	if len(zone.Backends) == 0 {
		zone.Backends = nil
	}
	marshal, err := jsonapi.Marshal(zone)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, binder.MIMEApplicationJSONApi, marshal)
}

func (r *ZoneRoute) Delete(c echo.Context) (err error) {
	err = r.db.Where("id = ?", c.Param("id")).Delete(&model.Zone{}).Error
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (r *ZoneRoute) Register(e *echo.Echo) {
	e.GET("/v1/zones/:id", r.Get)
	e.DELETE("/v1/zones/:id", r.Delete)
	e.POST("/v1/zones", r.Create)
	e.PATCH("/v1/zones/:id", r.Update)
	e.GET("/v1/zones", r.List)
}
