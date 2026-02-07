package merchant

import (
	"mime/multipart"
)

type MerchantRequest struct {
	Name       string               `json:"name"`
	Logo       multipart.File       `json:"logo"`
	LogoHeader multipart.FileHeader `json:"-"`
}

