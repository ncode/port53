package api

import (
	"github.com/DataDog/jsonapi"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/ncode/port53/pkg/binder"
	"github.com/ncode/port53/pkg/database"
	"github.com/spf13/viper"
)

func Server() {
	e := echo.New()

	switch viper.GetString("logLevel") {
	case "DEBUG":
		e.Debug = true
		e.Logger.SetLevel(log.DEBUG)
	case "INFO":
		e.Logger.SetLevel(log.INFO)
	case "WARN":
		e.Logger.SetLevel(log.WARN)
	case "ERROR":
		e.Logger.SetLevel(log.ERROR)
	case "OFF":
		e.Logger.SetLevel(log.OFF)
	}
	e.Binder = &binder.JsonApiBinder{}

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.DefaultLoggerConfig))

	db, err := database.Database()
	if err != nil {
		e.Logger.Fatal(err)
	}

	backend := &BackendRoute{db: db}
	backend.Register(e)
	zone := &ZoneRoute{db: db}
	zone.Register(e)

	e.Logger.Fatal(e.Start(viper.GetString("bindAddr")))
}

func JSONAPI(c echo.Context, code int, data interface{}) error {
	marshal, err := jsonapi.Marshal(data)
	if err != nil {
		return err
	}
	return c.Blob(code, binder.MIMEApplicationJSONApi, marshal)
}
