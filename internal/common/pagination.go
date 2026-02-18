package common

import (
	"encoding/json"
	"reflect"
)

type CursorPaginationResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasNext    bool   `json:"has_next"`
}

type PaginationMeta struct {
	TotalDocs     int64 `json:"total_docs"`
	Limit         int   `json:"limit"`
	TotalPages    int   `json:"total_pages"`
	Page          int   `json:"page"`
	PagingCounter int   `json:"paging_counter"`
	HasPrevPage   bool  `json:"has_prev_page"`
	HasNextPage   bool  `json:"has_next_page"`
	PrevPage      *int  `json:"prev_page,omitempty"`
	NextPage      *int  `json:"next_page,omitempty"`
}

type PaginatedResponse[T any] struct {
	Data T              `json:"docs"`
	Meta PaginationMeta `json:"meta"`
}

func (p PaginatedResponse[T]) MarshalJSON() ([]byte, error) {
	var data any = p.Data

	v := reflect.ValueOf(p.Data)
	if v.Kind() == reflect.Slice && v.IsNil() {
		data = reflect.MakeSlice(v.Type(), 0, 0).Interface()
	}

	return json.Marshal(&struct {
		Data any            `json:"docs"`
		Meta PaginationMeta `json:"meta"`
	}{
		Data: data,
		Meta: p.Meta,
	})
}

func BuildPaginationMeta(totalDocs int64, page, limit int) PaginationMeta {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	var totalPages int
	if totalDocs == 0 {
		totalPages = 0
		page = 1
	} else {
		totalPages = int((totalDocs + int64(limit) - 1) / int64(limit))
	}

	skip := (page - 1) * limit
	hasPrev := page > 1
	hasNext := page < totalPages

	var prevPage *int
	var nextPage *int

	if hasPrev {
		p := page - 1
		prevPage = &p
	}
	if hasNext {
		n := page + 1
		nextPage = &n
	}

	pagingCounter := 0
	if totalDocs > 0 {
		pagingCounter = skip + 1
	}

	return PaginationMeta{
		TotalDocs:     totalDocs,
		Limit:         limit,
		TotalPages:    totalPages,
		Page:          page,
		PagingCounter: pagingCounter,
		HasPrevPage:   hasPrev,
		HasNextPage:   hasNext,
		PrevPage:      prevPage,
		NextPage:      nextPage,
	}
}
