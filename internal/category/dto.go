package category

import (
	"lazeez-core/internal/common"
)

type CategoryRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type CategoryDTO struct {
	common.BaseDTO
	Name string `json:"name"`
	Icon string `json:"icon"`
}


type CategoryResponseSimplified struct {
	Name string `json:"name"`
	Icon string `json:"icon,omitempty"`
}
