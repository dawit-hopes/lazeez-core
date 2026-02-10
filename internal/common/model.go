package common

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Base struct {
	ID        string       `json:"id" db:"id"`
	IsDeleted bool         `json:"is_deleted" db:"is_deleted"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at,omitempty" db:"deleted_at"`
}

type Mappable interface {
	Table() string
	Columns() []string
	Values() []any
	Addr() []any
}

type Errors struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"err"`
}

func (e *Errors) Error() string {
	return fmt.Sprintf("code: %d, message: %s, error: %v", e.Code, e.Message, e.Err)
}

type Response struct {
	Data       any    `json:"data"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func NewResponse(data any, message string, statusCode int) *Response {
	return &Response{Data: data, Message: message, StatusCode: statusCode}
}

func (r *Response) ToJSON() []byte {
	json, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return json
}

func (b *Base) ToDTO() BaseDTO {
	return BaseDTO{
		ID:        b.ID,
		IsDeleted: b.IsDeleted,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
		DeletedAt: ToNullTimePtr(b.DeletedAt),
	}
}
