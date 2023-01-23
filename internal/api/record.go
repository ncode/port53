package api

import (
	"fmt"
	"net/http"

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
	err = r.db.Create(&record).Error
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: " {
			// Here it's more complex than in the backend or zone route, because
			// the record name is not unique, but the combination of name, type, zone and content
			// TOD: fix this
			var existingRecord model.Record
			err = r.db.First(&existingRecord, "name = ?", record.Name).Error
			if err != nil {
				return err
			}
			c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("%s/v1/records/%s", viper.GetString("serviceUrl"), existingRecord.ID))
			return c.String(http.StatusConflict, "Record already exists")
		} else if err.Error() == "UNIQUE constraint failed: recored.id" {
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
	err = r.db.Preload("Zone").Find(&records).Error
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
		return err
	}
	return JSONAPI(c, http.StatusOK, record)
}
