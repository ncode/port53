package api

import (
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
		//{
		//	name:                   "id conflict",
		//	input:                  `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZRE", "type": "records", "attributes": {"name": "internal.martinez.io", "type": "A", "ttl": 300, "content": "192.168.0.1"}, "relationships": { "zones": { "data": { "type": "zones", "id": "01F1ZQZJXQXZJXZJXZJXZJXZJX" }}}}}`,
		//	zoneInput:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
		//	expectedLocationHeader: fmt.Sprintf("%s/v1/records/%s", viper.GetString("serviceUrl"), "01F1ZQZJXQXZJXZJXZJXZJXZJX"),
		//	expectedStatusCode:     http.StatusConflict,
		//},
		//{
		//	name:                   "name conflict",
		//	input:                  `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZRE", "type": "records", "attributes": {"name": "internal.martinez.io", "type": "A", "ttl": 300, "content": "192.168.0.1"}, "relationships": { "zones": { "data": { "type": "zones", "id": "01F1ZQZJXQXZJXZJXZJXZJXZJX" }}}}}`,
		//	zoneInput:              `{"data": {"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX", "type": "zones", "attributes": {"name": "internal.martinez.io"}}}`,
		//	expectedLocationHeader: fmt.Sprintf("%s/v1/records/%s", viper.GetString("serviceUrl"), "01F1ZQZJXQXZJXZJXZJXZJXZJX"),
		//	expectedStatusCode:     http.StatusConflict,
		//},
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
