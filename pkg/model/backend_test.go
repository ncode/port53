package model

import (
	"testing"
	"time"

	"github.com/oklog/ulid/v2"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBackend_Get(t *testing.T) {
	tests := []struct {
		name          string
		backend       Backend
		preload       bool
		expectedError error
	}{
		{
			name: "Successful get with preload",
			backend: Backend{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			preload:       true,
			expectedError: nil,
		},
		{
			name: "Successful get without preload",
			backend: Backend{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			preload:       false,
			expectedError: nil,
		},
		{
			name: "Create with invalid ID",
			backend: Backend{
				ID:   "invalid",
				Name: ulid.Make().String(),
			},
			preload:       true,
			expectedError: ulid.ErrDataSize,
		},
		{
			name: "Get with non-existent ID",
			backend: Backend{
				Name: ulid.Make().String(),
			},
			preload:       true,
			expectedError: gorm.ErrRecordNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set up a test database and create a test zone
			db, err := gorm.Open(sqlite.Open("file:backend_model?mode=memory&cache=shared"), &gorm.Config{})
			if err != nil {
				t.Fatalf("Error setting up test database: %s", err)
			}
			err = db.AutoMigrate(&Backend{})
			if err != nil {
				t.Fatalf("Error running the migration: %s", err)
			}
			testBackend := Backend{
				ID:        test.backend.ID,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				Name:      test.backend.Name,
			}
			if err := db.Create(&testBackend).Error; err != nil {
				if test.backend.ID == "invalid" {
					if err == test.expectedError {
						return
					}
				}
				t.Fatalf("Error creating test zone: %s", err)
			}

			// Call the Get method and check the error
			err = test.backend.Get(db, test.preload)
			if err != test.expectedError {
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}

			if test.backend.ID == "" {
				return
			}

			// Check that the zone fields are correct
			if test.backend.ID != testBackend.ID {
				t.Errorf("Unexpected ID: got %s, want %s", test.backend.ID, testBackend.ID)
			}
			if test.backend.CreatedAt != testBackend.CreatedAt {
				t.Errorf("Unexpected CreatedAt: got %s, want %s", test.backend.CreatedAt, testBackend.CreatedAt)
			}
			if test.backend.UpdatedAt != testBackend.UpdatedAt {
				t.Errorf("Unexpected UpdatedAt: got %s, want %s", test.backend.UpdatedAt, testBackend.UpdatedAt)
			}
			if test.backend.Name != testBackend.Name {
				t.Errorf("Unexpected Name: got %s, want %s", test.backend.Name, testBackend.Name)
			}
			if test.preload && len(test.backend.Zones) != len(testBackend.Zones) {
				t.Errorf("Unexpected number of zones: got %d, want %d", len(test.backend.Zones), len(testBackend.Zones))
			}
		})
	}
}

func TestBackend_BeforeCreate(t *testing.T) {
	tests := []struct {
		name          string
		backend       Backend
		expectedError error
	}{
		{
			name: "Before create with empty ID",
			backend: Backend{
				ID: "",
			},
			expectedError: nil,
		},
		{
			name: "Before create with valid ID",
			backend: Backend{
				ID: "01D4K0Z5V5J9G5J5H5R5F5E5B5",
			},
			expectedError: nil,
		},
		{
			name: "Before create with invalid ID",
			backend: Backend{
				ID: "invalid",
			},
			expectedError: ulid.ErrDataSize,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Call the BeforeCreate method and check the error
			err := test.backend.BeforeCreate(nil)
			if err != nil {
				if err == test.expectedError {
					return
				}
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}

			// Check that the ID is set correctly
			if test.name == "Before create with empty ID" {
				_, err := ulid.Parse(test.backend.ID)
				if err != nil {
					t.Errorf("Unexpected error parsing ID: %s", err)
				}
			} else {
				if test.backend.ID != "01D4K0Z5V5J9G5J5H5R5F5E5B5" {
					t.Errorf("Unexpected ID: got %s, want %s", test.backend.ID, "01D4K0Z5V5J9G5J5H5R5F5E5B5")
				}
			}
		})
	}
}

func TestBackend_Delete(t *testing.T) {
	tests := []struct {
		name          string
		backend       Backend
		expectedError error
	}{
		{
			name: "Successful delete",
			backend: Backend{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			expectedError: nil,
		},
		{
			name: "Delete non-existent zone",
			backend: Backend{
				Name: ulid.Make().String(),
			},
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set up a test database and create a test zone
			db, err := gorm.Open(sqlite.Open("file:backend_model?mode=memory&cache=shared"), &gorm.Config{})
			if err != nil {
				t.Fatalf("Error setting up test database: %s", err)
			}
			err = db.AutoMigrate(&Backend{})
			if err != nil {
				t.Fatalf("Error running the migration: %s", err)
			}
			testBackend := Backend{
				ID:        test.backend.ID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Name:      test.backend.Name,
			}
			if err := db.Create(&testBackend).Error; err != nil {
				t.Fatalf("Error creating test zone: %s", err)
			}

			if test.backend.ID == "" {
				test.backend.ID = ulid.Make().String()
			}

			// Call the Delete method and check the error
			err = test.backend.Delete(db)
			if err != test.expectedError {
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}

			// Check that the zone is deleted
			var count int64
			db.Model(&Backend{}).Where("id = ?", test.backend.ID).Count(&count)
			if count != 0 {
				t.Errorf("Backend was not deleted")
			}
		})
	}
}

func TestBackend_Update(t *testing.T) {
	tests := []struct {
		name          string
		backend       Backend
		expectedError error
	}{
		{
			name: "Successful update",
			backend: Backend{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set up a test database and create a test backend
			db, err := gorm.Open(sqlite.Open("file:backend_model?mode=memory&cache=shared"), &gorm.Config{})
			if err != nil {
				t.Fatalf("Error setting up test database: %s", err)
			}
			err = db.AutoMigrate(&Backend{})
			if err != nil {
				t.Fatalf("Error running the migration: %s", err)
			}
			testBackend := Backend{
				ID:        test.backend.ID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Name:      test.backend.Name,
			}
			if err := db.Create(&testBackend).Error; err != nil {
				t.Fatalf("Error creating test backend: %s", err)
			}

			newBackend := Backend{Name: ulid.Make().String()}

			// Call the Update method and check the error
			err = test.backend.Update(db, newBackend)
			if err != test.expectedError {
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}

			// Check that the backend is updated
			updatedBackend := &Backend{ID: test.backend.ID}
			err = updatedBackend.Get(db, false)
			if err != err {
				t.Errorf("Unexpected error: got %s", err)
			}
			if updatedBackend.Name != newBackend.Name {
				t.Errorf("Unexpected Name: got %s, want %s", updatedBackend.Name, test.backend.Name)
			}
		})
	}
}

func TestBackend_AddZone(t *testing.T) {
	tests := []struct {
		name          string
		backend       Backend
		zone          Zone
		expectedError error
	}{
		{
			name: "Successful add",
			backend: Backend{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			zone: Zone{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set up a test database and create a test backend
			db, err := gorm.Open(sqlite.Open("file:backend_model?mode=memory&cache=shared"), &gorm.Config{})
			if err != nil {
				t.Fatalf("Error setting up test database: %s", err)
			}
			err = db.AutoMigrate(&Backend{}, &Zone{})
			if err != nil {
				t.Fatalf("Error running the migration: %s", err)
			}

			if err := db.Create(&test.backend).Error; err != nil {
				t.Fatalf("Error creating test backend: %s", err)
			}

			if err := db.Create(&test.zone).Error; err != nil {
				t.Fatalf("Error creating test zone: %s", err)
			}

			// Call the AddZone method and check the error
			err = test.backend.AddZone(db, &test.zone)
			if err != test.expectedError {
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}

			backend := &Backend{ID: test.backend.ID}
			err = backend.Get(db, true)
			if err != nil {
				t.Errorf("Unexpected error: got %s", err)
			}

			if backend.Zones[0].ID != test.zone.ID {
				t.Errorf("Unexpected zone: got %s, want %s", backend.Zones[0].ID, test.zone.ID)
			}
		})
	}
}

func TestBackend_RemoveZone(t *testing.T) {
	tests := []struct {
		name          string
		backend       Backend
		zone          Zone
		expectedError error
	}{
		{
			name: "Successful remove",
			backend: Backend{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			zone: Zone{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set up a test database and create a test backend
			db, err := gorm.Open(sqlite.Open("file:backend_model?mode=memory&cache=shared"), &gorm.Config{})
			if err != nil {
				t.Fatalf("Error setting up test database: %s", err)
			}
			err = db.AutoMigrate(&Backend{}, &Zone{})
			if err != nil {
				t.Fatalf("Error running the migration: %s", err)
			}

			if err := db.Create(&test.backend).Error; err != nil {
				t.Fatalf("Error creating test backend: %s", err)
			}

			if err := db.Create(&test.zone).Error; err != nil {
				t.Fatalf("Error creating test zone: %s", err)
			}

			if err := db.Model(&test.backend).Association("Zones").Append(&test.zone); err != nil {
				t.Fatalf("Error adding zone to backend: %s", err)
			}

			// Call the RemoveZone method and check the error
			err = test.backend.RemoveZone(db, &test.zone)
			if err != test.expectedError {
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}

			backend := &Backend{ID: test.backend.ID}
			err = backend.Get(db, true)
			if err != nil {
				t.Errorf("Unexpected error: got %s", err)
			}

			if len(backend.Zones) != 0 {
				t.Errorf("Unexpected zone: got %s, want %s", backend.Zones[0].ID, test.zone.ID)
			}
		})
	}
}

func TestBackend_ReplaceZones(t *testing.T) {
	tests := []struct {
		name          string
		backend       Backend
		zones         []*Zone
		expectedError error
	}{
		{
			name: "Successful replace",
			backend: Backend{
				ID:   ulid.Make().String(),
				Name: ulid.Make().String(),
			},
			zones: []*Zone{
				{
					ID:   ulid.Make().String(),
					Name: ulid.Make().String(),
				},
				{
					ID:   ulid.Make().String(),
					Name: ulid.Make().String(),
				},
			},
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set up a test database and create a test backend
			db, err := gorm.Open(sqlite.Open("file:backend_model?mode=memory&cache=shared"), &gorm.Config{})
			if err != nil {
				t.Fatalf("Error setting up test database: %s", err)
			}
			err = db.AutoMigrate(&Backend{}, &Zone{})
			if err != nil {
				t.Fatalf("Error running the migration: %s", err)
			}

			if err := db.Create(&test.backend).Error; err != nil {
				t.Fatalf("Error creating test backend: %s", err)
			}

			for _, zone := range test.zones {
				if err := db.Create(&zone).Error; err != nil {
					t.Fatalf("Error creating test zone: %s", err)
				}
			}

			// Call the ReplaceZones method and check the error
			err = test.backend.ReplaceZones(db, test.zones)
			if err != test.expectedError {
				t.Errorf("Unexpected error: got %s, want %s", err, test.expectedError)
			}

			backend := &Backend{ID: test.backend.ID}
			err = backend.Get(db, true)
			if err != nil {
				t.Errorf("Unexpected error: got %s", err)
			}

			if len(backend.Zones) != len(test.zones) {
				t.Errorf("Unexpected zone: got %s, want %s", backend.Zones[0].ID, test.zones[0].ID)
			}
		})
	}
}
