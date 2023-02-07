package model

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/oklog/ulid/v2"
)

func TestRecode_BeforeCreate(t *testing.T) {
	tests := []struct {
		name          string
		record        Record
		expectedError error
	}{
		{
			name: "Before create with empty ID",
			record: Record{
				ID: "",
			},
			expectedError: nil,
		},
		{
			name: "Before create with valid ID",
			record: Record{
				ID: "01D4K0Z5V5J9G5J5H5R5F5E5B5",
			},
			expectedError: nil,
		},
		{
			name: "Before create with invalid ID",
			record: Record{
				ID: "invalid",
			},
			expectedError: ulid.ErrDataSize,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Call the BeforeCreate method and check the error
			err := test.record.BeforeCreate(nil)
			if err != nil {
				if err == test.expectedError {
					return
				}
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}

			// Check that the ID is set correctly
			if test.name == "Before create with empty ID" {
				_, err := ulid.Parse(test.record.ID)
				if err != nil {
					t.Errorf("Unexpected error parsing ID: %s", err)
				}
			} else {
				if test.record.ID != "01D4K0Z5V5J9G5J5H5R5F5E5B5" {
					t.Errorf("Unexpected ID: got %s, want %s", test.record.ID, "01D4K0Z5V5J9G5J5H5R5F5E5B5")
				}
			}
		})
	}
}

func TestRecord_Delete(t *testing.T) {
	tests := []struct {
		name          string
		record        Record
		expectedError error
	}{
		{
			name: "Delete with valid ID",
			record: Record{
				ID: ulid.Make().String(),
			},
			expectedError: nil,
		},
		{
			name: "Delete with invalid ID",
			record: Record{
				ID: "invalid",
			},
			expectedError: nil,
		},
	}
	// Set up a test database and create a test zone
	db, err := gorm.Open(sqlite.Open("file:record?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Error setting up test database: %s", err)
	}
	err = db.AutoMigrate(&Record{})
	if err != nil {
		t.Fatalf("Error running the migration: %s", err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Call the Delete method and check the error
			err := test.record.Delete(db)
			if err != nil {
				if err == test.expectedError {
					return
				}
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}
		})
	}
}

func TestRecord_ReplaceZone(t *testing.T) {
	tests := []struct {
		name          string
		record        Record
		Zone          Zone
		expectedError error
	}{
		{
			name: "Update with valid ID",
			record: Record{
				ID: ulid.Make().String(),
			},
			Zone: Zone{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			expectedError: nil,
		},
	}
	// Set up a test database and create a test zone
	db, err := gorm.Open(sqlite.Open("file:record?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Error setting up test database: %s", err)
	}
	err = db.AutoMigrate(&Record{}, &Zone{})
	if err != nil {
		t.Fatalf("Error running the migration: %s", err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err = db.Create(&test.record).Error
			if err != nil {
				t.Fatalf("Error creating record: %s", err)
			}
			err = db.Create(&test.Zone).Error
			if err != nil {
				t.Fatalf("Error creating zone: %s", err)
			}

			// Call the UpdateZone method and check the error
			err := test.record.ReplaceZone(db, &test.Zone)
			if err != nil {
				if err == test.expectedError {
					return
				}
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}
			err = test.record.Get(db, true)
			if err != nil {
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}
			if test.record.ZoneID != test.Zone.ID {
				t.Errorf("Unexpected zone ID: got %s, want %s", test.record.ZoneID, test.Zone.ID)
			}
		})
	}
}

func TestRecord_Get(t *testing.T) {
	tests := []struct {
		name          string
		record        Record
		Zone          Zone
		expectedError error
	}{
		{
			name: "Get with valid ID",
			record: Record{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			Zone: Zone{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			expectedError: nil,
		},
	}
	// Set up a test database and create a test zone
	db, err := gorm.Open(sqlite.Open("file:record?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Error setting up test database: %s", err)
	}
	err = db.AutoMigrate(&Record{}, &Zone{})
	if err != nil {
		t.Fatalf("Error running the migration: %s", err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err = db.Create(&test.Zone).Error
			if err != nil {
				t.Fatalf("Error creating zone: %s", err)
			}
			test.record.ZoneID = test.Zone.ID
			err = db.Create(&test.record).Error
			if err != nil {
				t.Fatalf("Error creating record: %s", err)
			}

			err = test.record.Get(db, true)
			if err != nil {
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}
		})
	}
}
