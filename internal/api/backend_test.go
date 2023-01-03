package api

import (
	"github.com/DataDog/jsonapi"
	"github.com/ncode/trutinha/pkg/binder"
	"github.com/ncode/trutinha/pkg/database"
	"github.com/ncode/trutinha/pkg/model"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
)

func init() {
	viper.Set("database", "file::memory:?cache=shared")
}

var (
	backendResult  = &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "bind"}
	backendPayload = `{"data":{"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX","type":"backends","attributes":{"name":"bind"}}}`
)

func TestCreateBackend(t *testing.T) {
	e := echo.New()
	e.Binder = &binder.JsonApiBinder{}

	db, err := database.Database()
	if err != nil {
		panic(err)
	}

	routeBackend := &BackendRoute{db: db}
	c, recPost := postTestRequest("/v1/backends", backendPayload, e)
	err = routeBackend.Create(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, recPost.Code)
		assert.Equal(t, binder.MIMEApplicationJSONApi, recPost.Header().Get(echo.HeaderContentType))
		backend := &model.Backend{}
		assert.NoError(t, jsonapi.Unmarshal(recPost.Body.Bytes(), backend))
		assert.Equal(t, backendResult.Name, backend.Name)
		assert.Equal(t, backendResult.ID, backend.ID)
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
	c, recPost := postTestRequest("/v1/backends", backendPayload, e)
	err = routeBackend.Create(c)
	assert.NoError(t, err)

	c, recGet := getTestRequest("/v1/dc/:id", e)
	c.SetParamNames("id")
	c.SetParamValues(backendResult.ID)
	if assert.NoError(t, routeBackend.Get(c)) {
		assert.Equal(t, http.StatusOK, recGet.Code)
		assert.Equal(t, binder.MIMEApplicationJSONApi, recPost.Header().Get(echo.HeaderContentType))
		backend := &model.Backend{}
		assert.NoError(t, jsonapi.Unmarshal(recPost.Body.Bytes(), backend))
		assert.Equal(t, backendResult.Name, backend.Name)
		assert.Equal(t, backendResult.ID, backend.ID)
	}
}

func TestCreateBackendIDAlreadyExists(t *testing.T) {
	e := echo.New()
	e.Binder = &binder.JsonApiBinder{}

	db, err := database.Database()
	if err != nil {
		panic(err)
	}

	routeBackend := &BackendRoute{db: db}
	c, recPost := postTestRequest("/v1/backends", backendPayload, e)
	err = routeBackend.Create(c)
	assert.NoError(t, err)

	c, recFailPost := postTestRequest("/v1/backends", backendPayload, e)
	err = routeBackend.Create(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusConflict, recFailPost.Code)
		assert.Equal(t, binder.MIMEApplicationJSONApi, recPost.Header().Get(echo.HeaderContentType))
		assert.Contains(t, recFailPost.Header().Get(echo.HeaderLocation), backendResult.ID)
	}
}
