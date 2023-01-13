package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestParseQuery(t *testing.T) {
	testCases := []struct {
		name        string
		queryString string
		expected    *Query
	}{
		{
			name:        "parse filters",
			queryString: "filter[name]=Mary&filter[age]=25",
			expected: &Query{
				Filters: map[string][]string{
					"name": {"Mary"},
					"age":  {"25"},
				},
			},
		},
		{
			name:        "parse fields",
			queryString: "fields[articles]=title,body&fields[people]=name",
			expected: &Query{
				Includes: map[string]*Include{
					"articles": {Fields: []string{"title", "body"}},
					"people":   {Fields: []string{"name"}},
				},
			},
		},
		{
			name:        "parse include",
			queryString: "include=author",
			expected: &Query{
				Includes: map[string]*Include{
					"author": {},
				},
			},
		},
		{
			name:        "parse sort",
			queryString: "sort=age",
			expected: &Query{
				Sort: []string{"age"},
			},
		},
		{
			name:        "parse pagination",
			queryString: "page[limit]=20&page[offset]=10",
			expected: &Query{
				Page: &Page{
					Limit:  20,
					Offset: 10,
				},
			},
		},
		{
			name:        "parse multiple parameters",
			queryString: "filter[name]=Mary&filter[age]=25&page[limit]=20&page[offset]=10&include=author&fields[articles]=title,body&fields[people]=name&sort=age",
			expected: &Query{
				Filters: map[string][]string{
					"name": {"Mary"},
					"age":  {"25"},
				},
				Includes: map[string]*Include{
					"author":   {},
					"articles": {Fields: []string{"title", "body"}},
					"people":   {Fields: []string{"name"}},
				},
				Sort: []string{"age"},
				Page: &Page{
					Limit:  20,
					Offset: 10,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://example.com?"+tc.queryString, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := echo.New().NewContext(req, httptest.NewRecorder())
			query, err := ParseQuery(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(query.Filters) != len(tc.expected.Filters) {
				t.Errorf("expected filters %v but got %v", tc.expected.Filters, query.Filters)
			} else {
				for k, v := range query.Filters {
					if filters, ok := tc.expected.Filters[k]; !ok || !reflect.DeepEqual(filters, v) {
						t.Errorf("expected filters %v but got %v", tc.expected.Filters, query.Filters)
						break
					}
				}
			}
			if len(query.Includes) != len(tc.expected.Includes) {
				t.Errorf("expected includes %v but got %v", tc.expected.Includes, query.Includes)
			} else {
				for k, v := range query.Includes {
					if include, ok := tc.expected.Includes[k]; !ok || !reflect.DeepEqual(include, v) {
						t.Errorf("expected includes %v but got %v", tc.expected.Includes, query.Includes)
						break
					}
				}
			}
			if len(query.Sort) != len(tc.expected.Sort) {
				t.Errorf("expected sort %v but got %v", tc.expected.Sort, query.Sort)
			} else {
				for i := range query.Sort {
					if query.Sort[i] != tc.expected.Sort[i] {
						t.Errorf("expected sort %v but got %v", tc.expected.Sort, query.Sort)
						break
					}
				}
			}
			if !reflect.DeepEqual(query.Page, tc.expected.Page) {
				t.Errorf("expected page %v but got %v", tc.expected.Page, query.Page)
			}
		})
	}
}
