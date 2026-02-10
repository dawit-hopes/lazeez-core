package merchant

import (
	"lazeez-core/internal/common"
	"mime/multipart"
)

type MerchantRequest struct {
	Name       string               `json:"name"`
	Logo       multipart.File       `json:"logo"`
	LogoHeader multipart.FileHeader `json:"-"`
}

type MerchantDTO struct {
	common.BaseDTO
	Name string `json:"name"`
	Logo string `json:"logo"`
}
