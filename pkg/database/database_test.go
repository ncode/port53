package database

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

func TestDatabase(t *testing.T) {
	testCases := []struct {
		name     string
		config   string
		expected error
	}{
		{
			name:     "Successful connection",
			config:   "file::memory:?cache=shared",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			viper.Set("database", tc.config)
			db, err := Database()
			assert.Equal(t, tc.expected, err)
			if tc.expected == nil {
				assert.NotNil(t, db)
			} else {
				assert.Nil(t, db)
			}
		})
	}
}

func TestClose(t *testing.T) {
	testCases := []struct {
		name     string
		openDB   string
		expected error
	}{
		{
			name:     "Closing a closed connection",
			openDB:   "",
			expected: nil,
		},
		{
			name:     "Closing an open connection",
			openDB:   "file::memory:?cache=shared",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.openDB != "" {
				viper.Set("database", tc.openDB)
				_, err := gorm.Open(sqlite.Open(tc.openDB), &gorm.Config{})
				assert.NoError(t, err)
			}
			err := Close()
			assert.Equal(t, tc.expected, err)
		})
	}
}
