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

func TestRecordRoute_Create(t *testing.T) {
	defer TearDown()

	tests := []struct {
		name                   string
		input                  string
		zoneInput              string
		expectedData           *model.Record
		expectedLocationHeader string
		expectedStatusCode     int
	}{
		{
			name:               "valid input",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZRE", "type": "records", "attributes": {"name": "internal.martinez.io", "type": "A", "ttl": 300, "content": "192.168.0.1"}, "relationships": { "zones": { "data": { "type": "zones", "id": "01F1ZQZJXQXZJXZJXZJXZJXZJX" }}}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
			expectedData:       &model.Record{ID: "01F1ZQZJXQXZJXZJXZJXZJXZRE", Name: "internal.martinez.io", Type: "A", TTL: 300, Content: "192.168.0.1"},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "invalid input",
			input:              `{"data": {"type": "records", "attributes": {"name": ""}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:                   "conflict",
			input:                  `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZRE", "type": "records", "attributes": {"name": "internal.martinez.io", "type": "A", "ttl": 300, "content": "192.168.0.1"}, "relationships": { "zones": { "data": { "type": "zones", "id": "01F1ZQZJXQXZJXZJXZJXZJXZJX" }}}}}`,
			zoneInput:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
			expectedLocationHeader: fmt.Sprintf("%s/v1/records/%s", viper.GetString("serviceUrl"), "01F1ZQZJXQXZJXZJXZJXZJXZRE"),
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
			if test.zoneInput != "" {
				c, _ := postTestRequest("/v1/zones", test.zoneInput, e)
				err = routeZone.Create(c)
				assert.NoError(t, err)
			}

			routeRecord := &RecordRoute{db: db}
			c, rec := postTestRequest("/v1/records", test.input, e)
			err = routeRecord.Create(c)
			if assert.NoError(t, err) {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				if test.expectedData != nil {
					assert.Equal(t, binder.MIMEApplicationJSONApi, rec.Header().Get(echo.HeaderContentType))
					record := &model.Record{}
					assert.NoError(t, jsonapi.Unmarshal(rec.Body.Bytes(), record))
					assert.Equal(t, test.expectedData.Name, record.Name)
					assert.Equal(t, test.expectedData.ID, record.ID)
				}
				if test.expectedLocationHeader != "" {
					assert.Equal(t, test.expectedLocationHeader, rec.Header().Get(echo.HeaderLocation))
				}
			}
		})
	}
}

func TestRecordRoute_Get(t *testing.T) {
	defer TearDown()

	tests := []struct {
		name                   string
		id                     string
		input                  string
		zoneInput              string
		expectedData           *model.Record
		expectedLocationHeader string
		expectedStatusCode     int
	}{
		{
			name:               "valid input",
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZRE",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZRE", "type": "records", "attributes": {"name": "internal.martinez.io", "type": "A", "ttl": 300, "content": "192.168.0.1"}, "relationships": { "zones": { "data": { "type": "zones", "id": "01F1ZQZJXQXZJXZJXZJXZJXZJX" }}}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
			expectedData:       &model.Record{ID: "01F1ZQZJXQXZJXZJXZJXZJXZRE", Name: "internal.martinez.io", Type: "A", TTL: 300, Content: "192.168.0.1", ZoneID: "01F1ZQZJXQXZJXZJXZJXZJXZJX"},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "valid input",
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZNF",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZRE", "type": "records", "attributes": {"name": "internal.martinez.io", "type": "A", "ttl": 300, "content": "192.168.0.1"}, "relationships": { "zones": { "data": { "type": "zones", "id": "01F1ZQZJXQXZJXZJXZJXZJXZJX" }}}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
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
			if test.zoneInput != "" {
				c, _ := postTestRequest("/v1/zones", test.zoneInput, e)
				err = routeZone.Create(c)
				assert.NoError(t, err)
			}

			routeRecord := &RecordRoute{db: db}
			c, _ := postTestRequest("/v1/records", test.input, e)
			err = routeRecord.Create(c)
			assert.NoError(t, err)

			c, rec := getTestRequest("/v1/records/:id", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeRecord.Get(c)) {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				if test.expectedData != nil {
					assert.Equal(t, binder.MIMEApplicationJSONApi, rec.Header().Get(echo.HeaderContentType))
					record := &model.Record{}
					assert.NoError(t, jsonapi.Unmarshal(rec.Body.Bytes(), record))
					assert.Equal(t, test.expectedData.Name, record.Name)
					assert.Equal(t, test.expectedData.ID, record.ID)
					assert.Equal(t, test.expectedData.Type, record.Type)
					assert.Equal(t, test.expectedData.TTL, record.TTL)
					assert.Equal(t, test.expectedData.Content, record.Content)
					assert.Equal(t, test.expectedData.ZoneID, record.Zone.ID)
				}
			}
		})
	}
}

func TestRecordRoute_Delete(t *testing.T) {
	defer TearDown()

	tests := []struct {
		name                   string
		id                     string
		input                  string
		zoneInput              string
		expectedData           *model.Record
		expectedLocationHeader string
		expectedStatusCode     int
	}{
		{
			name:               "valid input",
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZRE",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZRE", "type": "records", "attributes": {"name": "internal.martinez.io", "type": "A", "ttl": 300, "content": "192.168.0.1"}, "relationships": { "zones": { "data": { "type": "zones", "id": "01F1ZQZJXQXZJXZJXZJXZJXZJX" }}}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
			expectedData:       &model.Record{ID: "01F1ZQZJXQXZJXZJXZJXZJXZRE", Name: "internal.martinez.io", Type: "A", TTL: 300, Content: "192.168.0.1", ZoneID: "01F1ZQZJXQXZJXZJXZJXZJXZJX"},
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
			if test.zoneInput != "" {
				c, _ := postTestRequest("/v1/zones", test.zoneInput, e)
				err = routeZone.Create(c)
				assert.NoError(t, err)
			}

			routeRecord := &RecordRoute{db: db}
			c, _ := postTestRequest("/v1/records", test.input, e)
			err = routeRecord.Create(c)
			assert.NoError(t, err)

			c, rec := getTestRequest("/v1/records/:id", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeRecord.Get(c)) {
				assert.Equal(t, http.StatusOK, rec.Code)
			}

			c, rec = deleteTestRequest("/v1/records", "", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeRecord.Delete(c)) {
				assert.Equal(t, http.StatusNoContent, rec.Code)
			}

			c, rec = getTestRequest("/v1/records/:id", e)
			c.SetParamNames("id")
			c.SetParamValues(test.id)
			if assert.NoError(t, routeRecord.Get(c)) {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
			}
		})
	}
}

func TestRecordRoute_List(t *testing.T) {
	defer TearDown()

	tests := []struct {
		name                   string
		id                     string
		input                  string
		zoneInput              string
		expectedData           []model.Record
		expectedLocationHeader string
		expectedStatusCode     int
	}{
		{
			name:               "valid input",
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZRE",
			input:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZRE", "type": "records", "attributes": {"name": "internal.martinez.io", "type": "A", "ttl": 300, "content": "192.168.0.1"}, "relationships": { "zones": { "data": { "type": "zones", "id": "01F1ZQZJXQXZJXZJXZJXZJXZJX" }}}}}`,
			zoneInput:          `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
			expectedData:       []model.Record{{ID: "01F1ZQZJXQXZJXZJXZJXZJXZRE", Name: "internal.martinez.io", Type: "A", TTL: 300, Content: "192.168.0.1", ZoneID: "01F1ZQZJXQXZJXZJXZJXZJXZJX"}},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "valid input",
			id:                 "01F1ZQZJXQXZJXZJXZJXZJXZNF",
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
			if test.zoneInput != "" {
				c, _ := postTestRequest("/v1/zones", test.zoneInput, e)
				err = routeZone.Create(c)
				assert.NoError(t, err)
			}

			routeRecord := &RecordRoute{db: db}
			if test.input != "" {
				c, _ := postTestRequest("/v1/records", test.input, e)
				err = routeRecord.Create(c)
				assert.NoError(t, err)
			}

			c, rec := getTestRequest("/v1/records", e)
			if assert.NoError(t, routeRecord.List(c)) {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				if len(test.expectedData) > 0 {
					assert.Equal(t, binder.MIMEApplicationJSONApi, rec.Header().Get(echo.HeaderContentType))
					var records []model.Record
					assert.NoError(t, jsonapi.Unmarshal(rec.Body.Bytes(), &records))
					assert.Equal(t, test.expectedData[0].Name, records[0].Name)
					assert.Equal(t, test.expectedData[0].ID, records[0].ID)
					assert.Equal(t, test.expectedData[0].Type, records[0].Type)
					assert.Equal(t, test.expectedData[0].TTL, records[0].TTL)
					assert.Equal(t, test.expectedData[0].Content, records[0].Content)
					assert.Equal(t, test.expectedData[0].ZoneID, records[0].Zone.ID)
				}
			}
		})
	}
}
