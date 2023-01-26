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

type BackendRoute struct {
	db *gorm.DB
}

// Create creates a new backend
func (r *BackendRoute) Create(c echo.Context) (err error) {
	var backend model.Backend
	if err := c.Bind(&backend); err != nil {
		return err
	}
	if backend.Name == "" {
		return c.String(http.StatusBadRequest, "Name is required")
	}
	err = r.db.Create(&backend).Error
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: backends.name" {
			var existingBackend model.Backend
			err = r.db.First(&existingBackend, "name = ?", backend.Name).Error
			if err != nil {
				return err
			}
			c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), existingBackend.ID))
			return c.String(http.StatusConflict, "Backend already exists")
		} else if err.Error() == "UNIQUE constraint failed: backends.id" {
			c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), backend.ID))
			return c.String(http.StatusConflict, "Backend already exists")
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return JSONAPI(c, http.StatusCreated, backend)
}

// List lists all backends
func (r *BackendRoute) List(c echo.Context) (err error) {
	var backends []model.Backend
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
		tx.Scopes(paginate(backends, p, tx)).Preload("Zones").Find(&backends)
		tx.Debug().Find(&backends)
	} else {
		r.db.Scopes(paginate(backends, p, r.db)).Preload("Zones").Find(&backends)
	}

	for pos, backend := range backends {
		if len(backend.Zones) == 0 {
			backends[pos].Zones = nil
		}
	}

	p.SetLinks(fmt.Sprintf("/v1/host?%s", query.BuildQuery()))
	return JSONAPIPaginated(c, http.StatusOK, backends, p.Link())
}

// Update updates a backend
func (r *BackendRoute) Update(c echo.Context) (err error) {
	backend := &model.Backend{ID: c.Param("id")}
	err = backend.Get(r.db, false)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Backend not found")
		}
		return err
	}
	newBackend := &model.Backend{}
	if err := c.Bind(newBackend); err != nil {
		return err
	}
	if (newBackend == nil) || (newBackend.Name == "") {
		return c.String(http.StatusBadRequest, "Name is required")
	}
	r.db.Model(&backend).Updates(&newBackend)
	return JSONAPI(c, http.StatusOK, backend)
}

// Get gets a backend
func (r *BackendRoute) Get(c echo.Context) (err error) {
	backend := &model.Backend{ID: c.Param("id")}
	err = backend.Get(r.db, true)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Backend not found")
		}
		return err
	}
	if len(backend.Zones) == 0 {
		backend.Zones = nil
	}
	return JSONAPI(c, http.StatusOK, backend)
}

// Delete deletes a backend
func (r *BackendRoute) Delete(c echo.Context) (err error) {
	backend := &model.Backend{ID: c.Param("id")}
	err = backend.Delete(r.db)
	if err != nil && err.Error() != "record not found" {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// GetZones gets all zones for a backend
func (r *BackendRoute) GetZones(c echo.Context) (err error) {
	backend := &model.Backend{ID: c.Param("id")}
	err = backend.Get(r.db, true)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Backend not found")
		}
		return err
	}
	if len(backend.Zones) == 0 {
		return c.String(http.StatusNotFound, "Backend doesn't have any zones")
	}
	return JSONAPI(c, http.StatusOK, backend.Zones)
}

// AddZone adds a zone to a backend
func (r *BackendRoute) AddZone(c echo.Context) (err error) {
	backend := &model.Backend{ID: c.Param("id")}
	err = backend.Get(r.db, false)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Backend not found")
		}
		return err
	}
	var zone model.Zone
	if err := c.Bind(&zone); err != nil {
		if strings.Contains(err.Error(), "body is not a json:api representation") {
			return c.String(http.StatusBadRequest, "Zone ID is required")
		}
		return err
	}
	if zone.ID == "" {
		return c.String(http.StatusBadRequest, "Zone ID is required")
	}
	existingZone := model.Zone{ID: zone.ID}
	err = existingZone.Get(r.db, false)
	if err != nil && err.Error() == "record not found" {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Zone not found")
		}
		return err
	}
	err = backend.AddZone(r.db, &existingZone)
	if err != nil {
		return err
	}
	return JSONAPI(c, http.StatusOK, existingZone)
}

// RemoveZone removes a zone from a backend
func (r *BackendRoute) RemoveZone(c echo.Context) (err error) {
	backend := &model.Backend{ID: c.Param("id")}
	err = backend.Get(r.db, false)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Backend not found")
		}
		return err
	}
	var zone model.Zone
	if err := c.Bind(&zone); err != nil {
		if strings.Contains(err.Error(), "body is not a json:api representation") {
			return c.String(http.StatusBadRequest, "Zone ID is required")
		}
		return err
	}
	if zone.ID == "" {
		return c.String(http.StatusBadRequest, "Zone ID is required")
	}
	err = backend.RemoveZone(r.db, &zone)
	if err != nil {
		return err
	}
	err = backend.Get(r.db, true)
	if err != nil {
		return err
	}
	if len(backend.Zones) == 0 {
		return c.String(http.StatusNoContent, "Backend doesn't have any zones")
	}
	return JSONAPI(c, http.StatusOK, backend.Zones)
}

// UpdateZones updates zones for a backend
func (r *BackendRoute) UpdateZones(c echo.Context) (err error) {
	backend := &model.Backend{ID: c.Param("id")}
	err = backend.Get(r.db, true)
	if err != nil {
		if err.Error() == "record not found" {
			return c.String(http.StatusNotFound, "Backend not found")
		}
		return err
	}
	var zones []model.Zone
	if err := c.Bind(&zones); err != nil {
		if strings.Contains(err.Error(), "body is not a json:api representation") {
			if body, err := io.ReadAll(c.Request().Body); err == nil {
				// This feel like a bug. Not 100% sure yet.
				if bytes.Equal(body, []byte("")) {
					err = r.db.Model(&backend).Association("Zones").Clear()
					if err != nil {
						return err
					}
					return c.String(http.StatusNoContent, "Removed all zones from backend")
				}
			}
		}
		return err
	}
	var ids []string
	for _, zone := range zones {
		ids = append(ids, zone.ID)
		if zone.ID == "" {
			return c.String(http.StatusBadRequest, "Zone ID is required")
		}
	}
	existingZones := make([]*model.Zone, 0)
	err = r.db.Find(&existingZones, "id IN (?)", ids).Error
	if err != nil {
		return err
	}
	if len(existingZones) == 0 || len(existingZones) != len(zones) {
		return c.String(http.StatusNotFound, "All zones must exist")
	}
	err = backend.ReplaceZones(r.db, existingZones)
	if err != nil {
		return err
	}
	return JSONAPI(c, http.StatusOK, existingZones)
}

// Register registers the routes for the backend
func (r *BackendRoute) Register(e *echo.Echo) {
	e.GET("/v1/backends/:id", r.Get)
	e.DELETE("/v1/backends/:id", r.Delete)
	e.POST("/v1/backends", r.Create)
	e.PATCH("/v1/backends/:id", r.Update)
	e.GET("/v1/backends", r.List)
	// Relationships
	e.GET("/v1/backends/:id/zones", r.GetZones)
	e.POST("/v1/backends/:id/zones", r.AddZone)
	e.PATCH("/v1/backends/:id/zones", r.UpdateZones)
	e.DELETE("/v1/backends/:id/zones", r.RemoveZone)
}
