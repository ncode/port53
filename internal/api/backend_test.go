package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/DataDog/jsonapi"
	"github.com/labstack/echo/v4"
	"github.com/ncode/port53/pkg/binder"
	"github.com/ncode/port53/pkg/database"
	"github.com/ncode/port53/pkg/model"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func init() {
	viper.Set("database", "file::memory:?cache=shared")
}

var (
	backendResult          = &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "bind"}
	backendPayload         = `{"data":{"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX","type":"backends","attributes":{"name":"bind"}}}`
	backendZoneResult      = &model.Zone{ID: "01GP0JQGFM61EKSCDGRDZ6H6QX", Name: "martinez.io"}
	backendZonePayload     = `{"data":{"id":"01GP0JQGFM61EKSCDGRDZ6H6QX","type":"zones","attributes":{"expire":604800,"minimum":3600,"mname":"@","name":"martinez.io","refresh":3600,"retry":600,"rname":"admin","serial":1,"ttl":3600}}}`
	backendZonePostPayload = `{"data":{"id":"01GP0JQGFM61EKSCDGRDZ6H6QX", "type":"zones"}}`
)

func TestCreateBackend(t *testing.T) {
	tests := []struct {
		name                   string
		input                  string
		expectedData           *model.Backend
		expectedLocationHeader string
		expectedStatusCode     int
	}{
		{
			name:               "valid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "backends", "attributes": {"name": "bind"}}}`,
			expectedData:       &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "bind"},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "invalid input",
			input:              `{"data": {"type": "backends", "attributes": {"name": ""}}}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:                   "id conflict",
			input:                  `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "backends", "attributes": {"name": "pdns"}}}`,
			expectedLocationHeader: fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), backendResult.ID),
			expectedStatusCode:     http.StatusConflict,
		},
		{
			name:                   "name conflict",
			input:                  `{"data": {"type": "backends", "attributes": {"name": "bind"}}}`,
			expectedLocationHeader: fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), backendResult.ID),
			expectedStatusCode:     http.StatusConflict,
		},
	}

	e := echo.New()
	e.Binder = &binder.JsonApiBinder{}

	db, err := database.Database()
	if err != nil {
		panic(err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			routeBackend := &BackendRoute{db: db}
			c, rec := postTestRequest("/v1/backends", test.input, e)
			err = routeBackend.Create(c)
			if assert.NoError(t, err) {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				if test.expectedData != nil {
					assert.Equal(t, binder.MIMEApplicationJSONApi, rec.Header().Get(echo.HeaderContentType))
					backend := &model.Backend{}
					assert.NoError(t, jsonapi.Unmarshal(rec.Body.Bytes(), backend))
					assert.Equal(t, test.expectedData.Name, backend.Name)
					assert.Equal(t, test.expectedData.ID, backend.ID)
				}
				if test.expectedLocationHeader != "" {
					assert.Equal(t, test.expectedLocationHeader, rec.Header().Get(echo.HeaderLocation))
				}
			}
		})
	}
}

func TestGetBackend(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		id                 string
		expectedData       *model.Backend
		expectedStatusCode int
	}{
		{
			name:               "record exists",
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "backends", "attributes": {"name": "bind"}}}`,
			expectedData:       &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "bind"},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "record does not exist",
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZZZ",
			expectedStatusCode: http.StatusNotFound,
		},
	}

	e := echo.New()
	e.Binder = &binder.JsonApiBinder{}

	db, err := database.Database()
	if err != nil {
		panic(err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			routeBackend := &BackendRoute{db: db}
			if test.input != "" {
				c, _ := postTestRequest("/v1/backends", test.input, e)
				err = routeBackend.Create(c)
				assert.NoError(t, err)
			}

			c, recGet := getTestRequest("/v1/backends/:id", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeBackend.Get(c)) {
				assert.Equal(t, test.expectedStatusCode, recGet.Code)
				if test.expectedData != nil {
					backend := &model.Backend{}
					assert.NoError(t, jsonapi.Unmarshal(recGet.Body.Bytes(), backend))
					assert.Equal(t, test.expectedData.Name, backend.Name)
					assert.Equal(t, test.expectedData.ID, backend.ID)
				}
			}
		})
	}
}

func TestDeleteBackend(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		id       string
		expected int
	}{
		{
			name:     "delete existing record",
			input:    `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZ00", "type": "backends", "attributes": {"name": "bind"}}}`,
			id:       "01F1ZQZJXQXZJXZJXZJXZJXZ00",
			expected: http.StatusCreated,
		},
	}
	e := echo.New()
	e.Binder = &binder.JsonApiBinder{}

	db, err := database.Database()
	if err != nil {
		panic(err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			routeBackend := &BackendRoute{db: db}
			c, _ := postTestRequest("/v1/backends", test.input, e)
			err = routeBackend.Create(c)
			assert.NoError(t, err)

			c, recGet := getTestRequest("/v1/backends/:id", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeBackend.Get(c)) {
				assert.Equal(t, http.StatusOK, recGet.Code)
			}

			c, recDelete := deleteTestRequest("/v1/backends/:id", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeBackend.Delete(c)) {
				assert.Equal(t, http.StatusNoContent, recDelete.Code)
			}

			c, recGet = getTestRequest("/v1/backends/:id", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeBackend.Get(c)) {
				assert.Equal(t, http.StatusNotFound, recGet.Code)
			}
		})
	}
}

func TestPatchBackendZone(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		zoneInput          string
		payload            string
		id                 string
		zoneID             string
		expectedData       *model.Zone
		expectedStatusCode int
	}{
		{
			name:               "valid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "backends", "attributes": {"name": "bind"}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			payload:            `{"data": [{"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "zones"}]}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			zoneID:             "01F1ZQZJXQXZJXZJXZJXZJZONE",
			expectedData:       &model.Zone{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "martinez.io"},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "backends", "attributes": {"name": "bind"}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			payload:            `{"data": [{"type": "zones"}]}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "nonexistent zone",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "backends", "attributes": {"name": "bind"}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			payload:            `{"data": [{"id":"01F1ZQZJXQXZJXZJXZJXZJLALA", "type": "zones"}]}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "remove zones from backend",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "backends", "attributes": {"name": "bind"}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			payload:            `{"data": []}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			expectedStatusCode: http.StatusNoContent,
		},
	}
	e := echo.New()
	e.Binder = &binder.JsonApiBinder{}

	db, err := database.Database()
	if err != nil {
		panic(err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			routeBackend := &BackendRoute{db: db}
			c, _ := postTestRequest("/v1/backends", test.input, e)
			err = routeBackend.Create(c)
			assert.NoError(t, err)

			routeZone := &ZoneRoute{db: db}
			c, _ = postTestRequest("/v1/zones", test.zoneInput, e)
			err = routeZone.Create(c)
			assert.NoError(t, err)

			c, recPatch := patchTestRequest("/v1/backends/:id/zones", test.payload, e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeBackend.UpdateZone(c)) {
				assert.Equal(t, test.expectedStatusCode, recPatch.Code)
				if test.zoneID != "" {
					var zones []model.Zone
					assert.NoError(t, jsonapi.Unmarshal(recPatch.Body.Bytes(), &zones))
					assert.Equal(t, test.zoneID, zones[0].ID)
				}

			}
		})
	}

}

func TestPostZoneBackend(t *testing.T) {
	e := echo.New()
	e.Binder = &binder.JsonApiBinder{}

	db, err := database.Database()
	if err != nil {
		panic(err)
	}

	routeBackend := &BackendRoute{db: db}
	c, _ := postTestRequest("/v1/backends", backendPayload, e)
	err = routeBackend.Create(c)
	assert.NoError(t, err)

	routeZone := &ZoneRoute{db: db}
	c, _ = postTestRequest("/v1/zones", backendZonePayload, e)
	err = routeZone.Create(c)
	assert.NoError(t, err)

	c, recPost := postTestRequest("/v1/backends/:id/zones", backendZonePostPayload, e)
	c.SetParamNames("id")
	c.SetParamValues(backendResult.ID)
	if assert.NoError(t, routeBackend.AddZone(c)) {
		assert.Equal(t, http.StatusOK, recPost.Code)
		var zone model.Zone
		assert.Equal(t, binder.MIMEApplicationJSONApi, recPost.Header().Get(echo.HeaderContentType))
		assert.NoError(t, jsonapi.Unmarshal(recPost.Body.Bytes(), &zone))
		assert.Equal(t, backendZoneResult.Name, zone.Name)
		assert.Equal(t, backendZoneResult.ID, zone.ID)
	}
}
