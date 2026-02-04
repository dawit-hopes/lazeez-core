package auth

import (
	"database/sql"
	"lazeez-core/internal/common"
)

type Role string

const (
	RoleAdmin         Role = "super admin"
	RoleBranchManager Role = "branch_manager"
)

type User struct {
	common.Base
	PhoneNumber     string         `json:"phone_number" db:"phone_number"`
	FullName        string         `json:"full_name" db:"full_name"`
	Password        string         `json:"-" db:"password"`
	Role            Role           `json:"role" db:"role"`
	BranchID        sql.NullString `json:"branch_id" db:"branch_id"`
	IsLocked        bool           `json:"is_locked" db:"is_locked"`
	IsFirstLogin    bool           `json:"is_first_login" db:"is_first_login"`
	LoggingAttempts int            `json:"logging_attempts" db:"logging_attempts"`
}

type AuthPayload struct {
	UserID   string `json:"uid"`
	BranchID string `json:"bid,omitempty"`
	Role     Role   `json:"rol"`
}

func (u *User) Table() string {
	return "users"
}

func (u *User) Columns() []string {
	return []string{"id", "user_name", "phone_number", "password", "role", "branch_id", "is_locked", "is_first_login", "logging_attempts", "created_at", "updated_at", "deleted_at", "is_deleted"}
}

func (u *User) Values() []any {
	return []any{u.ID, u.FullName, u.PhoneNumber, u.Password, u.Role, u.BranchID, u.IsLocked, u.IsFirstLogin, u.LoggingAttempts, u.CreatedAt, u.UpdatedAt, u.DeletedAt, u.IsDeleted}
}

func (u *User) Addr() []any {
	return []any{&u.ID, &u.FullName, &u.PhoneNumber, &u.Password, &u.Role, &u.BranchID, &u.IsLocked, &u.IsFirstLogin, &u.LoggingAttempts, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.IsDeleted}
}
