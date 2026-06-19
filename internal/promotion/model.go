package promotion

import (
	"lazeez-core/internal/common"
	"time"
)

type Promotion struct {
	common.Base
	MerchantID   string    `json:"merchant_id" db:"merchant_id"`
	Title        string    `json:"title" db:"title"`
	Description  string    `json:"description" db:"description"`
	BannerImage  string    `json:"banner_image" db:"banner_image"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	StartDate    time.Time `json:"start_date" db:"start_date"`
	EndDate      time.Time `json:"end_date" db:"end_date"`
}

func (p *Promotion) Table() string {
	return "promotions"
}

func (p *Promotion) Columns() []string {
	return []string{
		"id", "merchant_id", "title", "description", "banner_image",
		"is_active", "start_date", "end_date", "deleted_at", "is_deleted",
	}
}

func (p *Promotion) Values() []any {
	return []any{
		p.ID, p.MerchantID, p.Title, p.Description, p.BannerImage,
		p.IsActive, p.StartDate, p.EndDate, p.DeletedAt, p.IsDeleted,
	}
}

func (p *Promotion) Addr() []any {
	return []any{
		&p.ID, &p.MerchantID, &p.Title, &p.Description, &p.BannerImage,
		&p.IsActive, &p.StartDate, &p.EndDate, &p.DeletedAt, &p.IsDeleted,
		&p.CreatedAt, &p.UpdatedAt,
	}
}

func (p *Promotion) ToDTO() PromotionDTO {
	return PromotionDTO{
		BaseDTO: common.BaseDTO{
			ID:        p.ID,
			IsDeleted: p.IsDeleted,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			DeletedAt: common.ToNullTimePtr(p.DeletedAt),
		},
		MerchantID:  p.MerchantID,
		Title:       p.Title,
		Description: p.Description,
		BannerImage: p.BannerImage,
		IsActive:    p.IsActive,
		StartDate:   common.Date(p.StartDate),
		EndDate:     common.Date(p.EndDate),
	}
}

func (p *Promotion) ToPublicDTO() PromotionPublicDTO {
	return PromotionPublicDTO{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		BannerImage: p.BannerImage,
		StartDate:   common.Date(p.StartDate),
		EndDate:     common.Date(p.EndDate),
	}
}
