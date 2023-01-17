package model

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/oklog/ulid/v2"
)

func TestZone_BeforeCreate(t *testing.T) {
	tests := []struct {
		name          string
		zone          Zone
		expectedError error
	}{
		{
			name: "Before create with empty ID",
			zone: Zone{
				ID: "",
			},
			expectedError: nil,
		},
		{
			name: "Before create with valid ID",
			zone: Zone{
				ID: "01D4K0Z5V5J9G5J5H5R5F5E5B5",
			},
			expectedError: nil,
		},
		{
			name: "Before create with invalid ID",
			zone: Zone{
				ID: "invalid",
			},
			expectedError: ulid.ErrDataSize,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Call the BeforeCreate method and check the error
			err := test.zone.BeforeCreate(nil)
			if err != nil {
				if err == test.expectedError {
					return
				}
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}

			// Check that the ID is set correctly
			if test.name == "Before create with empty ID" {
				_, err := ulid.Parse(test.zone.ID)
				if err != nil {
					t.Errorf("Unexpected error parsing ID: %s", err)
				}
			} else {
				if test.zone.ID != "01D4K0Z5V5J9G5J5H5R5F5E5B5" {
					t.Errorf("Unexpected ID: got %s, want %s", test.zone.ID, "01D4K0Z5V5J9G5J5H5R5F5E5B5")
				}
			}
		})
	}
}

func TestZone_Delete(t *testing.T) {
	tests := []struct {
		name          string
		zone          Zone
		expectedError error
	}{
		{
			name: "Delete with valid ID",
			zone: Zone{
				ID: ulid.Make().String(),
			},
			expectedError: nil,
		},
		{
			name: "Delete with invalid ID",
			zone: Zone{
				ID: "invalid",
			},
			expectedError: nil,
		},
	}
	// Set up a test database and create a test zone
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Error setting up test database: %s", err)
	}
	err = db.AutoMigrate(&Zone{})
	if err != nil {
		t.Fatalf("Error running the migration: %s", err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Call the Delete method and check the error
			err := test.zone.Delete(db)
			if err != nil {
				if err == test.expectedError {
					return
				}
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}
		})
	}
}

func TestZone_Get(t *testing.T) {
	tests := []struct {
		name          string
		zone          Zone
		preload       bool
		expectedError error
	}{
		{
			name: "Successful get with preload",
			zone: Zone{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			preload:       true,
			expectedError: nil,
		},
		{
			name: "Successful get without preload",
			zone: Zone{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			preload:       false,
			expectedError: nil,
		},
		{
			name: "Create with invalid ID",
			zone: Zone{
				ID:   "invalid",
				Name: ulid.Make().String(),
			},
			preload:       true,
			expectedError: ulid.ErrDataSize,
		},
		{
			name: "Get with non-existent ID",
			zone: Zone{
				Name: ulid.Make().String(),
			},
			preload:       true,
			expectedError: gorm.ErrRecordNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set up a test database and create a test zone
			db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
			if err != nil {
				t.Fatalf("Error setting up test database: %s", err)
			}
			err = db.AutoMigrate(&Zone{}, &Record{})
			if err != nil {
				t.Fatalf("Error running the migration: %s", err)
			}
			testZone := Zone{
				ID:        test.zone.ID,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				Name:      test.zone.Name,
			}
			if err := db.Create(&testZone).Error; err != nil {
				if test.zone.ID == "invalid" {
					if err == test.expectedError {
						return
					}
				}
				t.Fatalf("Error creating test zone: %s", err)
			}

			// Call the Get method and check the error
			err = test.zone.Get(db, test.preload)
			if err != test.expectedError {
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}

			if test.zone.ID == "" {
				return
			}

			// Check that the zone fields are correct
			if test.zone.ID != testZone.ID {
				t.Errorf("Unexpected ID: got %s, want %s", test.zone.ID, testZone.ID)
			}
			if test.zone.CreatedAt != testZone.CreatedAt {
				t.Errorf("Unexpected CreatedAt: got %s, want %s", test.zone.CreatedAt, testZone.CreatedAt)
			}
			if test.zone.UpdatedAt != testZone.UpdatedAt {
				t.Errorf("Unexpected UpdatedAt: got %s, want %s", test.zone.UpdatedAt, testZone.UpdatedAt)
			}
			if test.zone.Name != testZone.Name {
				t.Errorf("Unexpected Name: got %s, want %s", test.zone.Name, testZone.Name)
			}
			if test.preload && len(test.zone.Backends) != len(testZone.Backends) {
				t.Errorf("Unexpected number of zones: got %d, want %d", len(test.zone.Backends), len(testZone.Backends))
			}
		})
	}
}
