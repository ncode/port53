package database

import (
	"time"

	"github.com/ncode/port53/pkg/model"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var database *gorm.DB

func Database() (db *gorm.DB, err error) {
	if database != nil {
		return database, err
	}

	database, err = gorm.Open(sqlite.Open(viper.GetString("database")), &gorm.Config{})
	if err != nil {
		return database, err
	}

	// Migrate the schema
	err = database.AutoMigrate(&model.Backend{}, &model.Zone{}, &model.Record{})
	if err != nil {
		return nil, err
	}

	if sqlDB, err := database.DB(); err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Minute * 10)
	}

	database.Exec(`PRAGMA foreign_keys = ON;`)
	database.Exec(`PRAGMA journal_mode=WAL;`)

	return database, err
}

func Close() error {
	if database == nil {
		return nil
	}

	sqlDB, err := database.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
