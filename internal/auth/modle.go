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
	return []string{"id", "full_name", "phone_number", "password", "role", "branch_id", "is_locked", "is_first_login", "logging_attempts", "deleted_at", "is_deleted"}
}

func (u *User) Values() []any {
	return []any{u.ID, u.FullName, u.PhoneNumber, u.Password, u.Role, u.BranchID, u.IsLocked, u.IsFirstLogin, u.LoggingAttempts, u.DeletedAt, u.IsDeleted}
}

func (u *User) Addr() []any {
	return []any{&u.ID, &u.FullName, &u.PhoneNumber, &u.Password, &u.Role, &u.BranchID, &u.IsLocked, &u.IsFirstLogin, &u.LoggingAttempts, &u.DeletedAt, &u.IsDeleted, &u.CreatedAt, &u.UpdatedAt}
}

func (u *User) ToDTO() UserDTO {
	return UserDTO{
		BaseDTO: common.BaseDTO{
			ID:        u.ID,
			IsDeleted: u.IsDeleted,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			DeletedAt: common.ToNullTimePtr(u.DeletedAt),
		},
		PhoneNumber:     u.PhoneNumber,
		FullName:        u.FullName,
		Role:            u.Role,
		BranchID:        u.BranchID.String,
		IsLocked:        u.IsLocked,
		IsFirstLogin:    u.IsFirstLogin,
		LoggingAttempts: u.LoggingAttempts,
	}
}
