package model

import (
	"testing"

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
