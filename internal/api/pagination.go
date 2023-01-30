package api

import (
	"fmt"
	"math"
	"strconv"

	"github.com/DataDog/jsonapi"
	"gorm.io/gorm"
)

type pagination struct {
	First    string      `json:"-"`
	Last     string      `json:"-"`
	Previous string      `json:"-"`
	Next     string      `json:"-"`
	Size     int         `json:"-"`
	Number   int         `json:"-"`
	Total    int64       `json:"-"`
	Type     string      `json:"-"`
	Self     string      `json:"-"`
	Data     interface{} `jsonapi:"primary,hosts"`
}

func (p *pagination) AtoiNumber(Number string) (err error) {
	if Number == "" {
		p.Number = 0
		return nil
	}
	p.Number, err = strconv.Atoi(Number)
	return err
}

func (p *pagination) AtoiSize(size string) (err error) {
	if size == "" {
		p.Size = 10
		return nil
	}
	p.Size, err = strconv.Atoi(size)
	return err
}

func (p *pagination) SetLinks(baseURL string) {
	totalPages := int(math.Ceil(float64(p.Total) / float64(p.Size)))

	if p.Number > 0 {
		p.Previous = fmt.Sprintf("%s?page[number]=%d&page[size]=%d", baseURL, p.Number-1, p.Size)
	}
	if p.Number < totalPages-1 {
		p.Next = fmt.Sprintf("%s?page[number]=%d&page[size]=%d", baseURL, p.Number+1, p.Size)
	}
	p.First = fmt.Sprintf("%s?page[number]=0&page[size]=%d", baseURL, p.Size)
	p.Last = fmt.Sprintf("%s?page[number]=%d&page[size]=%d", baseURL, totalPages-1, p.Size)
	if p.Number >= totalPages-1 {
		p.Self = fmt.Sprintf("%s?page[number]=%d&page[size]=%d", baseURL, totalPages-1, p.Size)
	} else {
		p.Self = fmt.Sprintf("%s?page[number]=%d&page[size]=%d", baseURL, p.Number, p.Size)
	}
}

func (p *pagination) Link() *jsonapi.Link {
	return &jsonapi.Link{
		First:    p.First,
		Last:     p.Last,
		Previous: p.Previous,
		Next:     p.Next,
		Self:     p.Self,
	}
}

func paginate(value interface{}, pages *pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	db.Model(value).Count(&pages.Total)

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pages.Number).Limit(pages.Size)
	}
}
