package common

import (
	"database/sql"
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
	return fmt.Sprintf("code: %s, message: %s, error: %v", e.Code, e.Message, e.Err)
}
