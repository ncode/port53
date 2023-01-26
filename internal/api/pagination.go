package api

import (
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
	if p.Number > 0 {
		p.Previous = baseURL + "?page[number]=" + strconv.Itoa(p.Number-1) + "&page[size]=" + strconv.Itoa(p.Size)
	}
	if p.Number < int(p.Total) {
		p.Next = baseURL + "?page[number]=" + strconv.Itoa(p.Number+1) + "&page[size]=" + strconv.Itoa(p.Size)
	}
	p.First = baseURL + "?page[number]=0&page[size]=" + strconv.Itoa(p.Size)
	p.Last = baseURL + "?page[number]=" + strconv.Itoa(int(p.Total)) + "&page[size]=" + strconv.Itoa(p.Size)
	p.Self = baseURL + "?page[number]=" + strconv.Itoa(p.Number) + "&page[size]=" + strconv.Itoa(p.Size)
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
