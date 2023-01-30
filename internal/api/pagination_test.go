package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type TestStruct struct {
	gorm.Model
	Name string
}

func TestAtoiNumber(t *testing.T) {
	p := &pagination{}

	err := p.AtoiNumber("")
	assert.Nil(t, err)
	assert.Equal(t, 0, p.Number)

	err = p.AtoiNumber("1")
	assert.Nil(t, err)
	assert.Equal(t, 1, p.Number)

	err = p.AtoiNumber("not_a_number")
	assert.NotNil(t, err)
}

func TestAtoiSize(t *testing.T) {
	p := &pagination{}

	err := p.AtoiSize("")
	assert.Nil(t, err)
	assert.Equal(t, 10, p.Size)

	err = p.AtoiSize("5")
	assert.Nil(t, err)
	assert.Equal(t, 5, p.Size)

	err = p.AtoiSize("not_a_number")
	assert.NotNil(t, err)
}

func TestSetLinks(t *testing.T) {
	baseURL := "https://example.com"
	tests := []struct {
		name     string
		paginate *pagination
		want     struct {
			first    string
			last     string
			previous string
			next     string
			self     string
		}
	}{
		{
			name: "first page with default size",
			paginate: &pagination{
				Number: 0,
				Size:   10,
				Total:  30,
			},
			want: struct {
				first    string
				last     string
				previous string
				next     string
				self     string
			}{
				first:    baseURL + "?page[number]=0&page[size]=10",
				last:     baseURL + "?page[number]=2&page[size]=10",
				previous: "",
				next:     baseURL + "?page[number]=1&page[size]=10",
				self:     baseURL + "?page[number]=0&page[size]=10",
			},
		},
		{
			name: "last page with custom size",
			paginate: &pagination{
				Number: 2,
				Size:   15,
				Total:  30,
			},
			want: struct {
				first    string
				last     string
				previous string
				next     string
				self     string
			}{
				first:    baseURL + "?page[number]=0&page[size]=15",
				last:     baseURL + "?page[number]=1&page[size]=15",
				previous: baseURL + "?page[number]=1&page[size]=15",
				next:     "",
				self:     baseURL + "?page[number]=1&page[size]=15",
			},
		},
		{
			name: "middle page with odd total count",
			paginate: &pagination{
				Number: 1,
				Size:   10,
				Total:  21,
			},
			want: struct {
				first    string
				last     string
				previous string
				next     string
				self     string
			}{
				first:    baseURL + "?page[number]=0&page[size]=10",
				last:     baseURL + "?page[number]=2&page[size]=10",
				previous: baseURL + "?page[number]=0&page[size]=10",
				next:     baseURL + "?page[number]=2&page[size]=10",
				self:     baseURL + "?page[number]=1&page[size]=10",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.paginate.SetLinks(baseURL)
			assert.Equal(t, tt.want.first, tt.paginate.First)
			assert.Equal(t, tt.want.last, tt.paginate.Last)
			assert.Equal(t, tt.want.previous, tt.paginate.Previous)
			assert.Equal(t, tt.want.next, tt.paginate.Next)
			assert.Equal(t, tt.want.self, tt.paginate.Self)
		})
	}
}

func TestLink(t *testing.T) {
	p := &pagination{
		Number:   2,
		Size:     10,
		Total:    100,
		First:    "http://localhost:8000/test?page[number]=0&page[size]=10",
		Last:     "http://localhost:8000/test?page[number]=99&page[size]=10",
		Previous: "http://localhost:8000/test?page[number]=1&page[size]=10",
		Next:     "http://localhost:8000/test?page[number]=3&page[size]=10",
		Self:     "http://localhost:8000/test?page[number]=2&page[size]=10",
	}

	link := p.Link()
	assert.Equal(t, p.First, link.First)
	assert.Equal(t, p.Last, link.Last)
	assert.Equal(t, p.Previous, link.Previous)
	assert.Equal(t, p.Next, link.Next)
	assert.Equal(t, p.Self, link.Self)
}
