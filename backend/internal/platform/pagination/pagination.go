package pagination

import "math"

const DefaultSize = 15

type Page struct {
	Number int
	Size   int
}

func New(page, size int) Page {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = DefaultSize
	}
	return Page{Number: page, Size: size}
}

func (p Page) Offset() int {
	return (p.Number - 1) * p.Size
}

func (p Page) TotalPages(total int) int {
	if total == 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(p.Size)))
}

type Response[T any] struct {
	Content       []T `json:"content"`
	Page          int `json:"page"`
	Size          int `json:"size"`
	TotalElements int `json:"totalElements"`
	TotalPages    int `json:"totalPages"`
}

func NewResponse[T any](items []T, total int, p Page) Response[T] {
	if items == nil {
		items = []T{}
	}
	return Response[T]{
		Content:       items,
		Page:          p.Number,
		Size:          p.Size,
		TotalElements: total,
		TotalPages:    p.TotalPages(total),
	}
}
