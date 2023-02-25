package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestParseQuery(t *testing.T) {
	testCases := []struct {
		name          string
		queryString   string
		expected      *Query
		expectedError string
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
			queryString: "page[size]=20&page[number]=10",
			expected: &Query{
				Page: &Page{
					Number: 10,
					Size:   20,
				},
			},
		},
		{
			name:          "parse pagination",
			queryString:   "page[size]=a&page[number]=1",
			expected:      nil,
			expectedError: "strconv.Atoi: parsing \"a\": invalid syntax",
		},
		{
			name:          "parse pagination",
			queryString:   "page[size]=1&page[number]=a",
			expected:      nil,
			expectedError: "strconv.Atoi: parsing \"a\": invalid syntax",
		},
		{
			name:          "parse pagination",
			queryString:   "%G",
			expected:      nil,
			expectedError: "invalid URL escape \"%G\"",
		},
		{
			name:        "parse multiple parameters",
			queryString: "filter[name]=Mary&filter[age]=25&page[size]=20&page[number]=10&include=author&fields[articles]=title,body&fields[people]=name&sort=age",
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
					Number: 10,
					Size:   20,
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
				if strings.Compare(err.Error(), tc.expectedError) == 0 {
					return
				}
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

func TestBuildQuery(t *testing.T) {
	testCases := []struct {
		name     string
		query    Query
		expected string
	}{
		{
			name: "empty query",
			query: Query{
				Filters:  make(map[string][]string),
				Includes: make(map[string]*Include),
				Sort:     []string{},
				Page:     nil,
			},
			expected: "",
		},
		{
			name: "filter query",
			query: Query{
				Filters: map[string][]string{
					"name": {"John", "Doe"},
					"age":  {"30"},
				},
				Includes: make(map[string]*Include),
				Sort:     []string{},
				Page:     nil,
			},
			expected: "filter[name]=John&filter[name]=Doe&filter[age]=30",
		},
		{
			name: "include query",
			query: Query{
				Filters: make(map[string][]string),
				Includes: map[string]*Include{
					"comments": {Fields: []string{}},
					"author":   {Fields: []string{}},
				},
				Sort: []string{},
				Page: nil,
			},
			expected: "include=comments,author",
		},
		{
			name: "fields query",
			query: Query{
				Filters: make(map[string][]string),
				Includes: map[string]*Include{
					"author": {Fields: []string{"name"}},
				},
				Sort: []string{},
				Page: nil,
			},
			expected: "include=author&fields[author]=name",
		},
		{
			name: "sort query",
			query: Query{
				Filters:  make(map[string][]string),
				Includes: make(map[string]*Include),
				Sort:     []string{"name", "-age"},
				Page:     nil,
			},
			expected: "sort=name,-age",
		},
		{
			name: "page query",
			query: Query{
				Filters:  make(map[string][]string),
				Includes: make(map[string]*Include),
				Sort:     []string{},
				Page:     &Page{Size: 10, Number: 2},
			},
			expected: "page[size]=10&page[number]=2",
		},
		{
			name: "all query",
			query: Query{
				Includes: map[string]*Include{
					"author": {Fields: []string{"id", "name"}},
				},
				Filters: map[string][]string{
					"title": {"Hello", "World"},
					"body":  {"Lorem", "Ipsum"},
				},
				Sort: []string{"id", "desc"},
				Page: &Page{Size: 10, Number: 2},
			},
			expected: "filter[title]=Hello&filter[title]=World&filter[body]=Lorem&filter[body]=Ipsum&include=author&fields[author]=id,name&sort=id,desc&page[size]=10&page[number]=2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.query.BuildQuery()
			if result != tc.expected {
				t.Errorf("BuildQuery() = %s; want %s", result, tc.expected)
			}
		})
	}
}
