package common

import (
	"database/sql"
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
