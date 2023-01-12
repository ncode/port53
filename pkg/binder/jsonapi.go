package binder

import (
	"io"
	"strings"

	"github.com/DataDog/jsonapi"
	"github.com/labstack/echo/v4"
)

const MIMEApplicationJSONApi string = "application/vnd.api+json"

type JsonApiBinder struct{}

func (j *JsonApiBinder) Bind(i interface{}, c echo.Context) (err error) {
	ctype := c.Request().Header.Get(echo.HeaderContentType)
	if ctype != MIMEApplicationJSONApi {
		return echo.ErrUnsupportedMediaType
	}

	// Use default binder if Content-Type is not application/vnd.api+json
	b := new(echo.DefaultBinder)
	err = b.Bind(i, c)
	if err != echo.ErrUnsupportedMediaType {
		return err
	}

	var body []byte
	body, err = io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	err = jsonapi.Unmarshal(body, i)
	if err != nil {
		if strings.Contains(err.Error(), "invalid character") {
			return echo.ErrBadRequest
		}
		return err
	}
	return err
}
