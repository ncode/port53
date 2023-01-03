package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/ncode/trutinha/pkg/binder"
	"github.com/spf13/viper"
)

func Server() {
	e := echo.New()
	e.Debug = viper.GetBool("debug")
	e.Binder = &binder.JsonApiBinder{}

	switch viper.GetString("logLevel") {
	case "DEBUG":
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

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.DefaultLoggerConfig))

	backend := &BackendRoute{}
	backend.Register(e)
	e.Logger.Fatal(e.Start(viper.GetString("bindAddr")))
}
