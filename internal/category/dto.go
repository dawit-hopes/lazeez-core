package category

import (
	"lazeez-core/internal/common"
	"mime/multipart"
)

type CategoryRequest struct {
	Name       string               `json:"name"`
	IconHeader multipart.FileHeader `json:"-"`
	Icon       multipart.File       `json:"icon"`
}

type CategoryDTO struct {
	common.BaseDTO
	Name string `json:"name"`
	Icon string `json:"icon"`
}
