package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
	"github.com/ncode/port53/pkg/model"
)

type ZoneRoute struct {
	db *gorm.DB
}

// Create creates a new zone
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
	return JSONAPI(c, http.StatusCreated, zone)
}

// List lists all zones
func (r *ZoneRoute) List(c echo.Context) (err error) {
	var zones []model.Zone

	query, err := ParseQuery(c)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid query parameters")
	}

	p := &pagination{Number: 0, Size: 10}
	if query.Page != nil {
		p = &pagination{Number: query.Page.Number, Size: query.Page.Size}
	}

	if len(query.Filters) > 0 {
		tx := r.db
		for filter, content := range query.Filters {
			for _, c := range content {
				tx = tx.Where(fmt.Sprintf("%s = ?", filter), c)
			}
		}
		err = tx.Scopes(paginate(zones, p, tx)).Preload("Backends").Preload("Records").Find(&zones).Error
	} else {
		err = r.db.Scopes(paginate(zones, p, r.db)).Preload("Backends").Preload("Records").Find(&zones).Error
	}
	if err != nil {
		return err
	}

	for pos, zone := range zones {
		if len(zone.Backends) == 0 {
			zones[pos].Backends = nil
		}
		if len(zone.Records) == 0 {
			zones[pos].Records = nil
		}
	}

	p.SetLinks(fmt.Sprintf("/v1/zones?%s", query.BuildQuery()))
	return JSONAPIPaginated(c, http.StatusOK, zones, p.Link())
}

// Update updates a zone
func (r *ZoneRoute) Update(c echo.Context) (err error) {
	zone := &model.Zone{ID: c.Param("id")}
	err = zone.Get(r.db, false)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Zone not found")
		}
		return err
	}
	newZone := &model.Zone{}
	if err := c.Bind(newZone); err != nil {
		return err
	}
	if (newZone == nil) || (newZone.Name == "") {
		return c.String(http.StatusBadRequest, "Name is required")
	}
	r.db.Model(&zone).Updates(&newZone)
	return JSONAPI(c, http.StatusOK, zone)
}

// Get gets a zone
func (r *ZoneRoute) Get(c echo.Context) (err error) {
	zone := model.Zone{ID: c.Param("id")}
	err = zone.Get(r.db, true)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Backend not found")
		}
		return err
	}
	if len(zone.Backends) == 0 {
		zone.Backends = nil
	}
	if len(zone.Records) == 0 {
		zone.Records = nil
	}
	return JSONAPI(c, http.StatusOK, zone)
}

// Delete deletes a zone
func (r *ZoneRoute) Delete(c echo.Context) (err error) {
	err = r.db.Where("id = ?", c.Param("id")).Delete(&model.Zone{}).Error
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// GetBackends gets a zone's backends
func (r *ZoneRoute) GetBackends(c echo.Context) (err error) {
	zone := model.Zone{ID: c.Param("id")}
	err = zone.Get(r.db, true)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Backend not found")
		}
		return err
	}
	if len(zone.Backends) == 0 {
		return JSONAPI(c, http.StatusNotFound, nil)
	}
	return JSONAPI(c, http.StatusOK, zone.Backends)
}

// AddBackend adds a zone to a backend
func (r *ZoneRoute) AddBackend(c echo.Context) (err error) {
	zone := model.Zone{ID: c.Param("id")}
	err = zone.Get(r.db, false)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Zone not found")
		}
		return err
	}
	var backend model.Backend
	if err := c.Bind(&backend); err != nil {
		if strings.Contains(err.Error(), "body is not a json:api representation") {
			return c.String(http.StatusBadRequest, "Backend ID is required")
		}
		return err
	}
	if backend.ID == "" {
		return c.String(http.StatusBadRequest, "Backend ID is required")
	}
	existingBackend := model.Backend{ID: backend.ID}
	err = existingBackend.Get(r.db, false)
	if err != nil && err.Error() == "record not found" {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Backend not found")
		}
		return err
	}
	err = zone.AddBackend(r.db, &existingBackend)
	if err != nil {
		return err
	}
	r.db.Find(&backend, "id = ?", backend.ID)
	return JSONAPI(c, http.StatusOK, backend)
}

// RemoveBackend removes a backend from a zone
func (r *ZoneRoute) RemoveBackend(c echo.Context) (err error) {
	zone := &model.Zone{ID: c.Param("id")}
	err = zone.Get(r.db, false)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Zone not found")
		}
		return err
	}
	var backend model.Backend
	if err := c.Bind(&backend); err != nil {
		if strings.Contains(err.Error(), "body is not a json:api representation") {
			return c.String(http.StatusBadRequest, "Backend ID is required")
		}
		return err
	}
	if backend.ID == "" {
		return c.String(http.StatusBadRequest, "Backend ID is required")
	}
	err = zone.RemoveBackend(r.db, &backend)
	if err != nil {
		return err
	}
	err = zone.Get(r.db, true)
	if err != nil {
		return err
	}
	if len(zone.Backends) == 0 {
		return c.String(http.StatusNoContent, "Zone doesn't have any backends")
	}
	return JSONAPI(c, http.StatusOK, zone.Backends)
}

// UpdateBackends updates backends for a zone
func (r *ZoneRoute) UpdateBackends(c echo.Context) (err error) {
	zone := &model.Zone{ID: c.Param("id")}
	err = zone.Get(r.db, true)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Backend not found")
		}
		return err
	}
	var backends []model.Backend
	if err := c.Bind(&backends); err != nil {
		if strings.Contains(err.Error(), "body is not a json:api representation") {
			if body, err := io.ReadAll(c.Request().Body); err == nil {
				// This feel like a bug. Not 100% sure yet.
				if bytes.Equal(body, []byte("")) {
					err = r.db.Model(&zone).Association("Backends").Clear()
					if err != nil {
						return err
					}
					return c.String(http.StatusNoContent, "Removed all backends from zone")
				}
			}
		}
		return err
	}

	var ids []string
	for _, backend := range backends {
		ids = append(ids, backend.ID)
		if backend.ID == "" {
			return c.String(http.StatusBadRequest, "Backend ID is required")
		}
	}
	existingBackends := make([]*model.Backend, 0)
	err = r.db.Find(&existingBackends, "id IN (?)", ids).Error
	if err != nil {
		return err
	}
	if len(existingBackends) == 0 || len(existingBackends) != len(backends) {
		return c.String(http.StatusNotFound, "All backends must exist")
	}
	err = zone.ReplaceBackends(r.db, existingBackends)
	if err != nil {
		return err
	}
	return JSONAPI(c, http.StatusOK, existingBackends)
}

// Register registers the routes
func (r *ZoneRoute) Register(e *echo.Echo) {
	e.GET("/v1/zones/:id", r.Get)
	e.DELETE("/v1/zones/:id", r.Delete)
	e.POST("/v1/zones", r.Create)
	e.PATCH("/v1/zones/:id", r.Update)
	e.GET("/v1/zones", r.List)
	// Relationships
	e.GET("/v1/zones/:id/backends", r.GetBackends)
	e.POST("/v1/zones/:id/backends", r.AddBackend)
	e.PATCH("/v1/backends/:id/backends", r.UpdateBackends)
	e.DELETE("/v1/zones/:id/backends", r.RemoveBackend)
}
