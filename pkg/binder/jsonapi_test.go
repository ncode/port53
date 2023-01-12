package binder

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestJsonApiBinder(t *testing.T) {
	type TestStruct struct {
		Field1 string `jsonapi:"primary,test"`
		Field2 int    `jsonapi:"attr,number"`
	}

	tests := []struct {
		name           string
		body           string
		contentType    string
		expectedStruct TestStruct
		expectedError  error
	}{
		{
			name: "valid jsonapi",
			body: `{
				"data": {
					"type": "test",
					"attributes": {
						"field1": "value1",
						"field2": 123
					}
				}
			}`,
			contentType:    MIMEApplicationJSONApi,
			expectedStruct: TestStruct{Field1: "value1", Field2: 123},
			expectedError:  nil,
		},
		{
			name:           "invalid content type",
			body:           `{"field1": "value1", "field2": 123}`,
			contentType:    echo.MIMEApplicationJSON,
			expectedStruct: TestStruct{},
			expectedError:  echo.ErrUnsupportedMediaType,
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(echo.POST, "/", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, tt.contentType)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			var testStruct TestStruct
			b := &JsonApiBinder{}
			err := b.Bind(&testStruct, c)

			if err != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if testStruct != tt.expectedStruct {
				t.Errorf("expected struct %v, got %v", tt.expectedStruct, testStruct)
			}
		})
	}
}
