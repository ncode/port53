package api

import (
	"fmt"
	"net/http"
	"strings"
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
	backendResult           = &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "bind"}
	backendPayload          = `{"data":{"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX","type":"backends","attributes":{"name":"bind"}}}`
	backendZoneResult       = &model.Zone{ID: "01GP0JQGFM61EKSCDGRDZ6H6QX", Name: "martinez.io"}
	backendZonePayload      = `{"data":{"id":"01GP0JQGFM61EKSCDGRDZ6H6QX","type":"zones","attributes":{"expire":604800,"minimum":3600,"mname":"@","name":"martinez.io","refresh":3600,"retry":600,"rname":"admin","serial":1,"ttl":3600}}}`
	backendZonePatchPayload = `{"data":[{"id":"01GP0JQGFM61EKSCDGRDZ6H6QX", "type":"zones"}]}`
	backendZonePostPayload  = `{"data":{"id":"01GP0JQGFM61EKSCDGRDZ6H6QX", "type":"zones"}}`
)

func TestCreateBackend(t *testing.T) {
	tests := []struct {
		name                   string
		input                  string
		expectedData           *model.Backend
		expectedLocationHeader string
		expected               int
	}{
		{
			name:         "valid input",
			input:        `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "backends", "attributes": {"name": "bind"}}}`,
			expectedData: &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "bind"},
			expected:     http.StatusCreated,
		},
		{
			name:     "invalid input",
			input:    `{"data": {"type": "backends", "attributes": {"name": ""}}}`,
			expected: http.StatusBadRequest,
		},
		{
			name:                   "id conflict",
			input:                  `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "backends", "attributes": {"name": "pdns"}}}`,
			expectedLocationHeader: fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), backendResult.ID),
			expected:               http.StatusConflict,
		},
		{
			name:                   "name conflict",
			input:                  `{"data": {"type": "backends", "attributes": {"name": "bind"}}}`,
			expectedLocationHeader: fmt.Sprintf("%s/v1/backends/%s", viper.GetString("serviceUrl"), backendResult.ID),
			expected:               http.StatusConflict,
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
			assert.NoError(t, err)
			assert.Equal(t, test.expected, rec.Code)
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
		})
	}
}

func TestGetBackend(t *testing.T) {
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

	c, recGet := getTestRequest("/v1/backends/:id", e)
	c.SetParamNames("id")
	c.SetParamValues(backendResult.ID)
	if assert.NoError(t, routeBackend.Get(c)) {
		assert.Equal(t, http.StatusOK, recGet.Code)
		assert.Equal(t, binder.MIMEApplicationJSONApi, recGet.Header().Get(echo.HeaderContentType))
		backend := &model.Backend{}
		assert.NoError(t, jsonapi.Unmarshal(recGet.Body.Bytes(), backend))
		assert.Equal(t, backendResult.Name, backend.Name)
		assert.Equal(t, backendResult.ID, backend.ID)
	}
}

func TestDeleteBackend(t *testing.T) {
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

	c, recDelete := deleteTestRequest("/v1/backends/:id", e)
	c.SetParamNames("id")
	c.SetParamValues(backendResult.ID)
	if assert.NoError(t, routeBackend.Delete(c)) {
		assert.Equal(t, http.StatusNoContent, recDelete.Code)
	}

	c, recGet := getTestRequest("/v1/backends/:id", e)
	c.SetParamNames("id")
	c.SetParamValues(backendResult.ID)
	if assert.NoError(t, routeBackend.Get(c)) {
		assert.Equal(t, http.StatusNotFound, recGet.Code)
	}
}

func TestPatchBackendNameEmpty(t *testing.T) {
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

	c, recPatch := patchTestRequest("/v1/backends/:id", strings.Replace(backendPayload, "bind", "", -1), e)
	c.SetParamNames("id")
	c.SetParamValues(backendResult.ID)
	if assert.NoError(t, routeBackend.Update(c)) {
		assert.Equal(t, http.StatusBadRequest, recPatch.Code)
	}
}

func TestPatchBackend(t *testing.T) {
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

	c, recPatch := patchTestRequest("/v1/backends/:id", strings.Replace(backendPayload, "bind", "powerdns", -1), e)
	c.SetParamNames("id")
	c.SetParamValues(backendResult.ID)
	if assert.NoError(t, routeBackend.Update(c)) {
		assert.Equal(t, http.StatusOK, recPatch.Code)
		backend := &model.Backend{}
		assert.Equal(t, binder.MIMEApplicationJSONApi, recPatch.Header().Get(echo.HeaderContentType))
		assert.NoError(t, jsonapi.Unmarshal(recPatch.Body.Bytes(), backend))
		assert.Equal(t, "powerdns", backend.Name)
		assert.Equal(t, backendResult.ID, backend.ID)
	}
}

func TestPatchZoneBackend(t *testing.T) {
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

	c, recPatch := patchTestRequest("/v1/backends/:id/zones", backendZonePatchPayload, e)
	c.SetParamNames("id")
	c.SetParamValues(backendResult.ID)
	if assert.NoError(t, routeBackend.UpdateZone(c)) {
		assert.Equal(t, http.StatusOK, recPatch.Code)
		var zones []model.Zone
		assert.Equal(t, binder.MIMEApplicationJSONApi, recPatch.Header().Get(echo.HeaderContentType))
		assert.NoError(t, jsonapi.Unmarshal(recPatch.Body.Bytes(), &zones))
		assert.Equal(t, 1, len(zones))
		assert.Equal(t, backendZoneResult.Name, zones[0].Name)
		assert.Equal(t, backendZoneResult.ID, zones[0].ID)
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
