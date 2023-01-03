package database

import (
	"github.com/ncode/trutinha/pkg/model"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
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

	database.Exec(`PRAGMA journal_mode=WAL;`)

	return database, err
}

func CleanupDatabase() (err error) {
	db, err := gorm.Open(sqlite.Open(viper.GetString("database")), &gorm.Config{})
	if err != nil {
		return err
	}

	db.Exec(`PRAGMA writable_schema = 1;
delete from sqlite_master where type in ('table', 'index', 'trigger');
PRAGMA writable_schema = 0;
VACUUM;
PRAGMA INTEGRITY_CHECK;`)

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	if err != nil {
		return err
	}

	return err
}
