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
	p := &pagination{
		Number: 2,
		Size:   10,
		Total:  100,
	}

	p.SetLinks("http://localhost:8000/test")

	assert.Equal(t, "http://localhost:8000/test?page[number]=1&page[size]=10", p.Previous)
	assert.Equal(t, "http://localhost:8000/test?page[number]=3&page[size]=10", p.Next)
	assert.Equal(t, "http://localhost:8000/test?page[number]=0&page[size]=10", p.First)
	assert.Equal(t, "http://localhost:8000/test?page[number]=100&page[size]=10", p.Last)
	assert.Equal(t, "http://localhost:8000/test?page[number]=2&page[size]=10", p.Self)
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
