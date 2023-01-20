package api

import (
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// Query is a struct that represents a query to the API
// https://jsonapi.org/format/#fetching
// https://jsonapi.org/format/#fetching-sparse-fieldsets
// https://jsonapi.org/format/#fetching-includes
// https://jsonapi.org/format/#query-parameters

type Query struct {
	Includes map[string]*Include
	Filters  map[string][]string
	Sort     []string
	Page     *Page
}

type Include struct {
	Fields []string
}

type Page struct {
	Offset int
	Limit  int
}

// ParseQuery parses a query string into a Query struct
func ParseQuery(c echo.Context) (*Query, error) {
	query := &Query{
		Includes: make(map[string]*Include),
		Filters:  make(map[string][]string),
	}

	qs := c.QueryString()
	for _, p := range strings.Split(qs, "&") {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		switch {
		case strings.HasPrefix(key, "filter["):
			// parse filters
			filterName := strings.TrimPrefix(key, "filter[")
			filterName = strings.TrimSuffix(filterName, "]")
			query.Filters[filterName] = append(query.Filters[filterName], value)

		case strings.HasPrefix(key, "fields["):
			// parse fields
			fieldName := strings.TrimPrefix(key, "fields[")
			fieldName = strings.TrimSuffix(fieldName, "]")
			if _, ok := query.Includes[fieldName]; !ok {
				query.Includes[fieldName] = &Include{}
			}
			query.Includes[fieldName].Fields = append(query.Includes[fieldName].Fields, strings.Split(value, ",")...)

		case key == "include":
			// parse includes
			for _, v := range strings.Split(value, ",") {
				if _, ok := query.Includes[v]; !ok {
					query.Includes[v] = &Include{}
				}
			}

		case key == "sort":
			// parse sorting
			query.Sort = strings.Split(value, ",")

		case key == "page[limit]":
			// parse page limit
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			if query.Page == nil {
				query.Page = &Page{}
			}
			query.Page.Limit = n
		case key == "page[offset]":
			// parse page offset
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			if query.Page == nil {
				query.Page = &Page{}
			}
			query.Page.Offset = n
		}
	}
	return query, nil
}
