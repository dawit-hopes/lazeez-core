package category

import (
	"lazeez-core/internal/common"
)

type CategoryRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
	// Station routes this category's items to "kitchen" or "bar". Optional.
	Station string `json:"station"`
}

type CategoryDTO struct {
	common.BaseDTO
	Name    string `json:"name"`
	Icon    string `json:"icon"`
	Station string `json:"station,omitempty"`
}


type CategoryResponseSimplified struct {
	Name    string `json:"name"`
	Icon    string `json:"icon,omitempty"`
	Station string `json:"station,omitempty"`
}
