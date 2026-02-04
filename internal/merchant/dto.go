package merchant

import "time"

type CreateMerchantRequest struct {
	Name string `json:"name"`
}

type UpdateMerchantRequest struct {
	Name string `json:"name"`
}

type GetMerchantResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
	IsDeleted bool      `json:"is_deleted"`
}
