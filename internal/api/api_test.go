package api

import (
	"net/http/httptest"
	"testing"

	"github.com/ncode/port53/pkg/binder"

	"github.com/labstack/echo/v4"
	"github.com/ncode/port53/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestJSONAPI(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		data     interface{}
		expected string
	}{
		{
			name:     "Test Case 1",
			code:     200,
			data:     []*model.Backend{{Name: "Backend 1", ID: "1"}, {Name: "Backend 2", ID: "2"}},
			expected: `{"data":[{"id":"1","type":"backends","attributes":{"name":"Backend 1"},"links":{"self":"/v1/backends/1"}},{"id":"2","type":"backends","attributes":{"name":"Backend 2"},"links":{"self":"/v1/backends/2"}}]}`,
		},
		{
			name:     "Test Case 2",
			code:     201,
			data:     &model.Backend{Name: "Backend 3", ID: "3"},
			expected: `{"data":{"id":"3","type":"backends","attributes":{"name":"Backend 3"},"links":{"self":"/v1/backends/3"}}}`,
		},
		{
			name:     "Test Case 3",
			code:     404,
			data:     nil,
			expected: `{"data":null}`,
		},
		{
			name:     "Test Case 4",
			code:     200,
			data:     &model.Backend{ID: "01F1ZQZJXQXZJXZJXZJXZJXZJX", Name: "bind"},
			expected: `{"data":{"id":"01F1ZQZJXQXZJXZJXZJXZJXZJX","type":"backends","attributes":{"name":"bind"},"links":{"self":"/v1/backends/01F1ZQZJXQXZJXZJXZJXZJXZJX"}}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(echo.GET, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := JSONAPI(c, test.code, test.data)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, rec.Body.String())
			assert.Equal(t, test.code, c.Response().Status)
			assert.Equal(t, binder.MIMEApplicationJSONApi, c.Response().Header().Get(echo.HeaderContentType))
		})
	}
}
