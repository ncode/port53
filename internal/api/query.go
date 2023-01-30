package api

import (
	"net/url"
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
	Size   int
	Number int
}

// ParseQuery parses a query string into a Query struct
func ParseQuery(c echo.Context) (*Query, error) {
	query := &Query{
		Includes: make(map[string]*Include),
		Filters:  make(map[string][]string),
	}

	qsEncoded := c.QueryString()
	qs, err := url.QueryUnescape(qsEncoded)
	if err != nil {
		return nil, err
	}

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

		case key == "page[size]":
			// parse page limit
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			if query.Page == nil {
				query.Page = &Page{}
			}
			query.Page.Size = n
		case key == "page[number]":
			// parse page offset
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			if query.Page == nil {
				query.Page = &Page{}
			}
			query.Page.Number = n
		}
	}
	return query, nil
}

func (q *Query) BuildQuery() (query string) {
	var b strings.Builder

	// Build filter parameters
	for k, v := range q.Filters {
		for _, vv := range v {
			b.WriteString("filter[" + k + "]=" + vv + "&")
		}
	}

	// Build include parameters
	if len(q.Includes) > 0 {
		b.WriteString("include=")
		for k := range q.Includes {
			b.WriteString(k + ",")
		}
		buff := b.String()
		b.Reset()
		b.WriteString(strings.TrimSuffix(buff, ",")) // remove trailing comma
		b.WriteString("&")
	}

	// Build field parameters
	for k, v := range q.Includes {
		if len(v.Fields) == 0 {
			continue
		}
		b.WriteString("fields[" + k + "]=")
		for _, vv := range v.Fields {
			b.WriteString(vv + ",")
		}
		buff := b.String()
		b.Reset()
		b.WriteString(strings.TrimSuffix(buff, ",")) // remove trailing comma
		b.WriteString("&")
	}

	// Build sort parameter
	if len(q.Sort) > 0 {
		b.WriteString("sort=")
		for _, v := range q.Sort {
			b.WriteString(v + ",")
		}
		buff := b.String()
		b.Reset()
		b.WriteString(strings.TrimSuffix(buff, ",")) // remove trailing comma
		b.WriteString("&")
	}

	// Build page parameters
	if q.Page != nil {
		b.WriteString("page[size]=" + strconv.Itoa(q.Page.Size) + "&")
		b.WriteString("page[number]=" + strconv.Itoa(q.Page.Number) + "&")
	}

	return strings.TrimSuffix(b.String(), "&")
}
