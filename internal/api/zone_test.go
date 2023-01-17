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

func TestZonedRoute_Create(t *testing.T) {
	tests := []struct {
		name                   string
		input                  string
		expectedData           *model.Zone
		expectedLocationHeader string
		expectedStatusCode     int
	}{
		{
			name:               "valid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			expectedData:       &model.Zone{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "martinez.io"},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "invalid input",
			input:              `{"data": {"type": "zones", "attributes": {"name": ""}}}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:                   "id conflict",
			input:                  `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "zones"}}}`,
			expectedLocationHeader: fmt.Sprintf("%s/v1/zones/%s", viper.GetString("serviceUrl"), "01F1ZQZJXQXZJXZJXZJXZJXZJX"),
			expectedStatusCode:     http.StatusConflict,
		},
		{
			name:                   "name conflict",
			input:                  `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXX0X", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			expectedLocationHeader: fmt.Sprintf("%s/v1/zones/%s", viper.GetString("serviceUrl"), "01F1ZQZJXQXZJXZJXZJXZJXZJX"),
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
			routeZone := &ZoneRoute{db: db}
			c, rec := postTestRequest("/v1/zones", test.input, e)
			err = routeZone.Create(c)
			if assert.NoError(t, err) {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				if test.expectedData != nil {
					assert.Equal(t, binder.MIMEApplicationJSONApi, rec.Header().Get(echo.HeaderContentType))
					zone := &model.Zone{}
					assert.NoError(t, jsonapi.Unmarshal(rec.Body.Bytes(), zone))
					assert.Equal(t, test.expectedData.Name, zone.Name)
					assert.Equal(t, test.expectedData.ID, zone.ID)
				}
				if test.expectedLocationHeader != "" {
					assert.Equal(t, test.expectedLocationHeader, rec.Header().Get(echo.HeaderLocation))
				}
			}
		})
	}
}

func TestZonedRoute_Get(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		id                 string
		expectedData       *model.Zone
		expectedStatusCode int
	}{
		{
			name:               "record exists",
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			expectedData:       &model.Zone{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "martinez.io"},
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
			routeZone := &ZoneRoute{db: db}
			if test.input != "" {
				c, _ := postTestRequest("/v1/zones", test.input, e)
				err = routeZone.Create(c)
				assert.NoError(t, err)
			}

			c, recGet := getTestRequest("/v1/zones/:id", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeZone.Get(c)) {
				assert.Equal(t, test.expectedStatusCode, recGet.Code)
				if test.expectedData != nil {
					zone := &model.Zone{}
					assert.NoError(t, jsonapi.Unmarshal(recGet.Body.Bytes(), zone))
					assert.Equal(t, test.expectedData.Name, zone.Name)
					assert.Equal(t, test.expectedData.ID, zone.ID)
				}
			}
		})
	}
}

func TestZoneRoute_Delete(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		id       string
		expected int
	}{
		{
			name:     "delete existing record",
			input:    `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZ00", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
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
			routeZone := &ZoneRoute{db: db}
			c, _ := postTestRequest("/v1/zones", test.input, e)
			err = routeZone.Create(c)
			assert.NoError(t, err)

			c, recGet := getTestRequest("/v1/zones/:id", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeZone.Get(c)) {
				assert.Equal(t, http.StatusOK, recGet.Code)
			}

			c, recDelete := deleteTestRequest("/v1/zones/:id", "", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeZone.Delete(c)) {
				assert.Equal(t, http.StatusNoContent, recDelete.Code)
			}

			c, recGet = getTestRequest("/v1/zones/:id", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeZone.Get(c)) {
				assert.Equal(t, http.StatusNotFound, recGet.Code)
			}
		})
	}
}

func TestZoneRoute_List(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedData       []model.Zone
		expectedStatusCode int
	}{
		{
			name:               "valid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			expectedData:       []model.Zone{{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "martinez.io"}},
			expectedStatusCode: http.StatusOK,
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
			route := &ZoneRoute{db: db}
			c, _ := postTestRequest("/v1/zones", test.input, e)
			err = route.Create(c)
			assert.NoError(t, err)

			c, rec := getTestRequest("/v1/zones", e)
			if assert.NoError(t, route.List(c)) {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				var zones []model.Zone
				assert.NoError(t, jsonapi.Unmarshal(rec.Body.Bytes(), &zones))
				assert.Equal(t, test.expectedData[0].ID, zones[0].ID)
				assert.Equal(t, test.expectedData[0].Name, zones[0].Name)
			}
		})
	}
}

func TestZonedRoute_GetBackend(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		zoneInput          string
		payload            string
		id                 string
		zoneID             string
		expectedData       *model.Backend
		expectedStatusCode int
	}{
		{
			name:               "valid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "nsd"}}}`,
			payload:            `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends"}}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			zoneID:             "01F1ZQZJXQXZJXZJXZJXZJZONE",
			expectedData:       &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJZONE", Name: "nsd"},
			expectedStatusCode: http.StatusOK,
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
			routeZone := &ZoneRoute{db: db}
			c, _ := postTestRequest("/v1/zones", test.input, e)
			err = routeZone.Create(c)
			assert.NoError(t, err)

			routeBackend := &BackendRoute{db: db}
			c, _ = postTestRequest("/v1/backends", test.zoneInput, e)
			err = routeBackend.Create(c)
			assert.NoError(t, err)

			c, _ = postTestRequest("/v1/zones/:id/backends", test.payload, e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			assert.NoError(t, routeZone.AddBackend(c))

			c, recGet := getTestRequest("/v1/zones/:id/backends", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeZone.GetBackends(c)) {
				assert.Equal(t, http.StatusOK, recGet.Code)
				var zones []model.Zone
				assert.NoError(t, jsonapi.Unmarshal(recGet.Body.Bytes(), &zones))
				assert.Equal(t, test.expectedData.ID, zones[0].ID)
				assert.Equal(t, test.expectedData.Name, zones[0].Name)
			}
		})
	}
}
