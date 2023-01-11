package api

type Query struct {
	Filter  map[string]interface{} `query:"filter"`
	Page    map[string]interface{} `query:"page"`
	include string                 `query:"include"`
	fields  map[string]string      `query:"fields"`
}
