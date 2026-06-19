package promotion

import (
	"lazeez-core/internal/common"
	"mime/multipart"
)

type PromotionRequest struct {
	Title             string               `json:"title"`
	Description       string               `json:"description"`
	BannerImageHeader multipart.FileHeader `json:"-"`
	BannerImage       multipart.File       `json:"banner_image"`
	IsActive          *bool                `json:"is_active"`
	StartDate         common.Date          `json:"start_date"`
	EndDate           common.Date          `json:"end_date"`
	MerchantID        string               `json:"merchant_id"`
}

func (r *PromotionRequest) IsEmpty() bool {
	return r.Title == "" &&
		r.Description == "" &&
		r.BannerImage == nil &&
		r.IsActive == nil &&
		r.StartDate.IsZero() &&
		r.EndDate.IsZero()
}

type PromotionDTO struct {
	common.BaseDTO
	MerchantID  string      `json:"merchant_id"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	BannerImage string      `json:"banner_image"`
	IsActive    bool        `json:"is_active"`
	StartDate   common.Date `json:"start_date"`
	EndDate     common.Date `json:"end_date"`
}

type PromotionPublicDTO struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	BannerImage string      `json:"banner_image"`
	StartDate   common.Date `json:"start_date"`
	EndDate     common.Date `json:"end_date"`
}
