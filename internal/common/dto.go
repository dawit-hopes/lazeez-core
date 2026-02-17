package common

import "time"

type BaseDTO struct {
	ID        string     `json:"id"`
	IsDeleted bool       `json:"is_deleted"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type Filter struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Search string `json:"search"`
}

type PaginationResponse struct {
	Data        []*any  `json:"data"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
	Total       int  `json:"total"`
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
}
