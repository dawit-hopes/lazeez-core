package feedback

import "lazeez-core/internal/common"

type RatingRequest struct {
	Rating      int      `json:"rating"`
	Comment     string   `json:"comment"`
	Tags        []string `json:"tags"`
	PhoneNumber string   `json:"phone_number"`
}

type StayRatingRequest struct {
	Rating      int      `json:"rating"`
	Comment     string   `json:"comment"`
	Tags        []string `json:"tags"`
	PhoneNumber string   `json:"phone_number"`
	TableName   string   `json:"table_name"`
}

type OrderRatingDTO struct {
	common.BaseDTO
	OrderID     string   `json:"order_id"`
	OrderNumber int      `json:"order_number,omitempty"`
	BranchID    string   `json:"branch_id,omitempty"`
	SessionKey  string   `json:"session_key,omitempty"`
	Rating      int      `json:"rating"`
	Comment     string   `json:"comment,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	PhoneNumber string   `json:"phone_number,omitempty"`
}

type StayRatingDTO struct {
	common.BaseDTO
	SessionKey  string   `json:"session_key,omitempty"`
	BranchID    string   `json:"branch_id,omitempty"`
	TableName   string   `json:"table_name,omitempty"`
	Rating      int      `json:"rating"`
	Comment     string   `json:"comment,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	PhoneNumber string   `json:"phone_number,omitempty"`
}
