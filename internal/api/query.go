package api

// Query is a struct that represents a query to the API
// https://jsonapi.org/format/#fetching
// https://jsonapi.org/format/#fetching-sparse-fieldsets
// https://jsonapi.org/format/#fetching-includes
// https://jsonapi.org/format/#query-parameters
type Query struct {
	Filter  map[string]string `query:"filter"`
	Page    map[string]int    `query:"page"`
	include string            `query:"include"`
	sort    string            `query:"sort"`
}
