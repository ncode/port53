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
	viper.Set("database", "file:zone?mode=memory&cache=shared")
}

func TestZonedRoute_Create(t *testing.T) {
	defer TearDown()

	tests := []struct {
		name                   string
		input                  string
		expectedData           *model.Zone
		expectedLocationHeader string
		expectedStatusCode     int
	}{
		{
			name:               "valid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
			expectedData:       &model.Zone{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "internal.martinez.io"},
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
			input:                  `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXX0X", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
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
	defer TearDown()

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
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
			expectedData:       &model.Zone{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "internal.martinez.io", Records: []*model.Record{}},
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
					fmt.Println(recGet.Body.String())
					assert.NoError(t, jsonapi.Unmarshal(recGet.Body.Bytes(), zone))
					assert.Equal(t, test.expectedData.Name, zone.Name)
					assert.Equal(t, test.expectedData.ID, zone.ID)
				}
			}
		})
	}
}

func TestZoneRoute_Delete(t *testing.T) {
	defer TearDown()

	tests := []struct {
		name     string
		input    string
		id       string
		expected int
	}{
		{
			name:     "delete existing record",
			input:    `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZ11", "type": "zones", "attributes": {"name": "another.martinez.io"}}}`,
			id:       "01F1ZQZJXQXZJXZJXZJXZJXZ11",
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

func TestZoneRoute_Update(t *testing.T) {
	defer TearDown()

	tests := []struct {
		name               string
		input              string
		payload            string
		id                 string
		expectedData       *model.Zone
		expectedStatusCode int
	}{
		{
			name:               "valid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			payload:            `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "home.martinez.io"}}}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			expectedData:       &model.Zone{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "home.martinez.io"},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			payload:            `{"data": {"type": "zones"}}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			expectedStatusCode: http.StatusBadRequest,
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

			c, recPatch := patchTestRequest("/v1/zones/:id", test.payload, e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeZone.Update(c)) {
				assert.Equal(t, test.expectedStatusCode, recPatch.Code)
				if test.expectedData != nil {
					var zone model.Zone
					assert.NoError(t, jsonapi.Unmarshal(recPatch.Body.Bytes(), &zone))
					assert.Equal(t, test.expectedData.ID, zone.ID)
				}

			}
		})
	}
}

func TestZoneRoute_List(t *testing.T) {
	defer TearDown()

	tests := []struct {
		name               string
		input              string
		expectedData       []model.Zone
		expectedStatusCode int
	}{
		{
			name:               "valid input",
			input:              `{"data": {"id":"01GQ0MJ5N2X42FB43WC25XDE1A", "type": "zones", "attributes": {"name": "external.martinez.io"}}}`,
			expectedData:       []model.Zone{{ID: "01GQ0MJ5N2X42FB43WC25XDE1A", Name: "external.martinez.io"}},
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
	defer TearDown()

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
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJBACK", "type": "backends", "attributes": {"name": "nsd"}}}`,
			payload:            `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJBACK", "type": "backends"}}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			zoneID:             "01F1ZQZJXQXZJXZJXZJXZJZBACK",
			expectedData:       &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJBACK", Name: "nsd"},
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
				var backends []model.Backend
				assert.NoError(t, jsonapi.Unmarshal(recGet.Body.Bytes(), &backends))
				assert.Equal(t, test.expectedData.ID, backends[0].ID)
				assert.Equal(t, test.expectedData.Name, backends[0].Name)
			}
		})
	}
}

func TestZoneRoute_AddBackend(t *testing.T) {
	defer TearDown()

	tests := []struct {
		name               string
		input              string
		backendInput       string
		payload            string
		id                 string
		backendID          string
		expectedData       *model.Backend
		expectedStatusCode int
	}{
		{
			name:               "valid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			backendInput:       `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "powerdns"}}}`,
			payload:            `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends"}}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			backendID:          "01F1ZQZJXQXZJXZJXZJXZJZONE",
			expectedData:       &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "powerdns"},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			backendInput:       `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "powerdns"}}}`,
			payload:            `{"data": {"type": "backends"}}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "nonexistent backend",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			backendInput:       `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "powerdns"}}}`,
			payload:            `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJLALA", "type": "backends"}}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "remove backend from zone",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			backendInput:       `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "powerdns"}}}`,
			payload:            `{"data": null}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			expectedStatusCode: http.StatusBadRequest,
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
			c, _ = postTestRequest("/v1/backends", test.backendInput, e)
			err = routeBackend.Create(c)
			assert.NoError(t, err)

			c, recPost := postTestRequest("/v1/zones/:id/backends", test.payload, e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeZone.AddBackend(c)) {
				assert.Equal(t, test.expectedStatusCode, recPost.Code)
				if test.backendID != "" {
					var backend model.Backend
					assert.NoError(t, jsonapi.Unmarshal(recPost.Body.Bytes(), &backend))
					assert.Equal(t, test.backendID, backend.ID)
				}

			}
		})
	}
}

func TestZoneRoute_RemoveBackend(t *testing.T) {
	defer TearDown()

	tests := []struct {
		name               string
		input              string
		backendInput       string
		payload            string
		deletePayload      string
		id                 string
		BackendID          string
		expectedData       *model.Backend
		expectedStatusCode int
	}{
		{
			name:               "delete without any lasting zone",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			backendInput:       `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "nsd"}}}`,
			payload:            `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends"}}`,
			deletePayload:      `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends"}}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			BackendID:          "01F1ZQZJXQXZJXZJXZJXZJZONE",
			expectedData:       &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJZONE", Name: "nsd"},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "delete with zone left",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			backendInput:       `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "nsd"}}}`,
			payload:            `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends"}}`,
			deletePayload:      `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJlALA", "type": "backends"}}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			BackendID:          "01F1ZQZJXQXZJXZJXZJXZJZONE",
			expectedData:       &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJZONE", Name: "nsd"},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			backendInput:       `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "nsd"}}}`,
			payload:            `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends"}}`,
			deletePayload:      `{"data": null }`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			BackendID:          "01F1ZQZJXQXZJXZJXZJXZJZONE",
			expectedData:       &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJZONE", Name: "nsd"},
			expectedStatusCode: http.StatusBadRequest,
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
			c, _ = postTestRequest("/v1/backends", test.backendInput, e)
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
				var backends []model.Backend
				assert.NoError(t, jsonapi.Unmarshal(recGet.Body.Bytes(), &backends))
				assert.Equal(t, test.expectedData.ID, backends[0].ID)
				assert.Equal(t, test.expectedData.Name, backends[0].Name)
			}

			c, recDelete := deleteTestRequest("/v1/zones/:id/backends", test.deletePayload, e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeZone.RemoveBackend(c)) {
				assert.Equal(t, test.expectedStatusCode, recDelete.Code)
			}

			c, recGet = getTestRequest("/v1/zones/:id/backends", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeZone.GetBackends(c)) {
				if test.expectedStatusCode == http.StatusNoContent {
					assert.Equal(t, http.StatusNotFound, recGet.Code)
				} else {
					assert.Equal(t, http.StatusOK, recGet.Code)
				}
			}
		})
	}
}

func TestZoneRoute_UpdateBackends(t *testing.T) {
	defer TearDown()

	tests := []struct {
		name               string
		input              string
		zoneInput          string
		payload            string
		id                 string
		backendID          string
		expectedData       *model.Backend
		expectedStatusCode int
	}{
		{
			name:               "valid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "bind"}}}`,
			payload:            `{"data": [{"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends"}]}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			backendID:          "01F1ZQZJXQXZJXZJXZJXZJZONE",
			expectedData:       &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "bind"},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "bind"}}}`,
			payload:            `{"data": [{"type": "backends"}]}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "nonexistent zone",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "bind"}}}`,
			payload:            `{"data": [{"id":"01F1ZQZJXQXZJXZJXZJXZJLALA", "type": "backends"}]}`,
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZJX",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "remove zones from backend",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "martinez.io"}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJZONE", "type": "backends", "attributes": {"name": "bind"}}}`,
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
			routeZone := &ZoneRoute{db: db}
			c, _ := postTestRequest("/v1/zones", test.input, e)
			err = routeZone.Create(c)
			assert.NoError(t, err)

			routeBackend := &BackendRoute{db: db}
			c, _ = postTestRequest("/v1/backends", test.zoneInput, e)
			err = routeBackend.Create(c)
			assert.NoError(t, err)

			c, recPatch := patchTestRequest("/v1/zones/:id/backends", test.payload, e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeZone.UpdateBackends(c)) {
				assert.Equal(t, test.expectedStatusCode, recPatch.Code)
				if test.backendID != "" {
					var backends []model.Backend
					assert.NoError(t, jsonapi.Unmarshal(recPatch.Body.Bytes(), &backends))
					assert.Equal(t, test.backendID, backends[0].ID)
				}
			}
		})
	}
}
