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
