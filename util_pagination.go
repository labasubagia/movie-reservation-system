package main

import (
	"math"
	"reflect"
)

const (
	Page     = 1
	PageSize = 10

	MaxPageSize = 50
)

type Paginate[T any] struct {
	Items       []T   `json:"items"`
	TotalItems  int64 `json:"total_items"`
	TotalPage   int64 `json:"total_page"`
	CurrentPage int64 `json:"current_page"`
	PageSize    int64 `json:"page_size"`
}

func NewPaginate[T any](items []T, total, page, size int64) *Paginate[T] {
	p := Paginate[T]{
		Items:       []T{},
		CurrentPage: page,
		PageSize:    size,
		TotalPage:   GetTotalPage(total, size),
		TotalItems:  total,
	}

	v := reflect.ValueOf(items)
	if (v.Kind() == reflect.Array || v.Kind() == reflect.Slice) && v.Len() > 0 {
		p.Items = items
	}
	return &p
}

func GetTotalPage(total, pageSize int64) int64 {
	return int64(math.Ceil(float64(total) / float64(pageSize)))
}

type PaginateInput struct {
	Page int64
	Size int64
}

func NewPaginateInput(page, size int64) PaginateInput {
	p := PaginateInput{
		Page: Page,
		Size: PageSize,
	}
	if page > 0 {
		p.Page = page
	}
	if size > 0 && size <= MaxPageSize {
		p.Size = size
	}
	return p
}
