package feedback

import (
	"database/sql"
	"strings"

	"lazeez-core/internal/common"

	"github.com/lib/pq"
)

type OrderRating struct {
	common.Base
	OrderID     string         `json:"order_id" db:"order_id"`
	BranchID    string         `json:"branch_id" db:"branch_id"`
	SessionKey  string         `json:"session_key" db:"session_key"`
	Rating      int            `json:"rating" db:"rating"`
	Comment     sql.NullString `json:"comment" db:"comment"`
	Tags        pq.StringArray `json:"tags" db:"tags"`
	PhoneNumber sql.NullString `json:"phone_number" db:"phone_number"`
	OrderNumber int            `json:"order_number" db:"order_number"`
}

func (r *OrderRating) Table() string {
	return "order_ratings"
}

func (r *OrderRating) Columns() []string {
	return []string{
		"id", "order_id", "branch_id", "session_key", "rating", "comment", "tags", "phone_number",
		"deleted_at", "is_deleted",
	}
}

func (r *OrderRating) Values() []any {
	return []any{
		r.ID, r.OrderID, r.BranchID, r.SessionKey, r.Rating, r.Comment, r.Tags, r.PhoneNumber,
		r.DeletedAt, r.IsDeleted,
	}
}

func (r *OrderRating) Addr() []any {
	return []any{
		&r.ID, &r.OrderID, &r.BranchID, &r.SessionKey, &r.Rating, &r.Comment, &r.Tags, &r.PhoneNumber,
		&r.DeletedAt, &r.IsDeleted, &r.CreatedAt, &r.UpdatedAt,
	}
}

func (r *OrderRating) ToDTO() OrderRatingDTO {
	dto := OrderRatingDTO{
		BaseDTO: common.BaseDTO{
			ID:        r.ID,
			IsDeleted: r.IsDeleted,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			DeletedAt: common.ToNullTimePtr(r.DeletedAt),
		},
		OrderID:     r.OrderID,
		OrderNumber: r.OrderNumber,
		BranchID:    r.BranchID,
		SessionKey:  r.SessionKey,
		Rating:      r.Rating,
		Tags:        []string(r.Tags),
	}
	if r.Comment.Valid {
		dto.Comment = r.Comment.String
	}
	if r.PhoneNumber.Valid {
		dto.PhoneNumber = r.PhoneNumber.String
	}
	if dto.Tags == nil {
		dto.Tags = []string{}
	}
	return dto
}

type StayRating struct {
	common.Base
	SessionKey  string         `json:"session_key" db:"session_key"`
	BranchID    string         `json:"branch_id" db:"branch_id"`
	TableName   sql.NullString `json:"table_name" db:"table_name"`
	Rating      int            `json:"rating" db:"rating"`
	Comment     sql.NullString `json:"comment" db:"comment"`
	Tags        pq.StringArray `json:"tags" db:"tags"`
	PhoneNumber sql.NullString `json:"phone_number" db:"phone_number"`
}

func (r *StayRating) Table() string {
	return "stay_ratings"
}

func (r *StayRating) Columns() []string {
	return []string{
		"id", "session_key", "branch_id", "table_name", "rating", "comment", "tags", "phone_number",
		"deleted_at", "is_deleted",
	}
}

func (r *StayRating) Values() []any {
	return []any{
		r.ID, r.SessionKey, r.BranchID, r.TableName, r.Rating, r.Comment, r.Tags, r.PhoneNumber,
		r.DeletedAt, r.IsDeleted,
	}
}

func (r *StayRating) Addr() []any {
	return []any{
		&r.ID, &r.SessionKey, &r.BranchID, &r.TableName, &r.Rating, &r.Comment, &r.Tags, &r.PhoneNumber,
		&r.DeletedAt, &r.IsDeleted, &r.CreatedAt, &r.UpdatedAt,
	}
}

func (r *StayRating) ToDTO() StayRatingDTO {
	dto := StayRatingDTO{
		BaseDTO: common.BaseDTO{
			ID:        r.ID,
			IsDeleted: r.IsDeleted,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			DeletedAt: common.ToNullTimePtr(r.DeletedAt),
		},
		SessionKey: r.SessionKey,
		BranchID:   r.BranchID,
		Rating:     r.Rating,
		Tags:       []string(r.Tags),
	}
	if r.TableName.Valid {
		dto.TableName = r.TableName.String
	}
	if r.Comment.Valid {
		dto.Comment = r.Comment.String
	}
	if r.PhoneNumber.Valid {
		dto.PhoneNumber = r.PhoneNumber.String
	}
	if dto.Tags == nil {
		dto.Tags = []string{}
	}
	return dto
}

func nullStringFromOptional(value string) sql.NullString {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func tagsFromRequest(tags []string) pq.StringArray {
	if len(tags) == 0 {
		return pq.StringArray{}
	}
	out := make(pq.StringArray, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			out = append(out, tag)
		}
	}
	return out
}
