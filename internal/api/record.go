package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/ncode/port53/pkg/model"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type RecordRoute struct {
	db *gorm.DB
}

// Create creates a new record
func (r *RecordRoute) Create(c echo.Context) (err error) {
	var record model.Record
	if err := c.Bind(&record); err != nil {
		return err
	}
	if record.Name == "" {
		return c.String(http.StatusBadRequest, "Name is required")
	}
	if record.Zone == nil {
		return c.String(http.StatusBadRequest, "Zone is required")
	}
	record.ZoneID = record.Zone.ID
	err = r.db.Create(&record).Error
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: ") {
			c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("%s/v1/records/%s", viper.GetString("serviceUrl"), record.ID))
			return c.String(http.StatusConflict, "Record already exists")
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return JSONAPI(c, http.StatusCreated, record)
}

// List lists all records
func (r *RecordRoute) List(c echo.Context) (err error) {
	var records []model.Record
	err = r.db.Find(&records).Error
	for _, record := range records {
		if err := record.Get(r.db, true); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return JSONAPI(c, http.StatusOK, records)
}

// Get gets a record
func (r *RecordRoute) Get(c echo.Context) (err error) {
	record := model.Record{ID: c.Param("id")}
	err = record.Get(r.db, true)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Record not found")
		}
		return err
	}
	return JSONAPI(c, http.StatusOK, record)
}

// Register registers the routes
func (r *RecordRoute) Register(e *echo.Echo) {
	e.GET("/v1/records/:id", r.Get)
	//e.DELETE("/v1/records/:id", r.Delete)
	e.POST("/v1/records", r.Create)
	//e.PATCH("/v1/records/:id", r.Update)
	e.GET("/v1/records", r.List)
	// Relationships
	//e.GET("/v1/records/:id/zones", r.GetZone)
	//e.PATCH("/v1/records/:id/zones", r.UpdateZone)
	//e.DELETE("/v1/records/:id/zones", r.RemoveZone)
}
