package api

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/ncode/port53/pkg/database"

	"github.com/labstack/echo/v4"
	"github.com/ncode/port53/pkg/binder"
)

func TearDown() {
	database.Close()
}

func getTestRequest(target string, e *echo.Echo) (c echo.Context, recGet *httptest.ResponseRecorder) {
	get := httptest.NewRequest(http.MethodGet, target, nil)
	get.Header.Set(echo.HeaderContentType, binder.MIMEApplicationJSONApi)
	recGet = httptest.NewRecorder()
	return e.NewContext(get, recGet), recGet
}

func deleteTestRequest(target string, payload string, e *echo.Echo) (c echo.Context, recDelete *httptest.ResponseRecorder) {
	var del *http.Request
	if payload != "" {
		del = httptest.NewRequest(http.MethodDelete, target, strings.NewReader(payload))
	} else {
		del = httptest.NewRequest(http.MethodDelete, target, nil)
	}
	del.Header.Set(echo.HeaderContentType, binder.MIMEApplicationJSONApi)
	recDelete = httptest.NewRecorder()
	return e.NewContext(del, recDelete), recDelete
}

func patchTestRequest(target string, payload string, e *echo.Echo) (c echo.Context, recPatch *httptest.ResponseRecorder) {
	patch := httptest.NewRequest(http.MethodPatch, target, strings.NewReader(payload))
	patch.Header.Set(echo.HeaderContentType, binder.MIMEApplicationJSONApi)
	recPatch = httptest.NewRecorder()
	return e.NewContext(patch, recPatch), recPatch
}

func postTestRequest(target string, payload string, e *echo.Echo) (c echo.Context, recPost *httptest.ResponseRecorder) {
	post := httptest.NewRequest(http.MethodPost, target, strings.NewReader(payload))
	post.Header.Set(echo.HeaderContentType, binder.MIMEApplicationJSONApi)
	recPost = httptest.NewRecorder()
	return e.NewContext(post, recPost), recPost
}
