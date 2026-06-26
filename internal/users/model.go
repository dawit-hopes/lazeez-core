package users

import (
	"database/sql"
	"lazeez-core/internal/common"
)

type Role string

const (
	RoleAdmin              Role = "super_admin"
	RoleBranchManager      Role = "branch_manager"
	RoleSuperBranchManager Role = "super_branch_admin"
	RoleFrontDeskAgent     Role = "front_desk_agent"
	RoleRoomServiceStaff   Role = "room_service_staff"
	RoleWaiter             Role = "waiter"
	RoleKitchenStaff       Role = "kitchen_staff"
	RoleBarista            Role = "barista"
	RoleCashier            Role = "cashier"
)

type User struct {
	common.Base
	PhoneNumber     string         `json:"phone_number" db:"phone_number"`
	FullName        string         `json:"full_name" db:"full_name"`
	Password        string         `json:"-" db:"password"`
	PasscodeHash    sql.NullString `json:"-" db:"passcode_hash"`
	Role            Role           `json:"role" db:"role"`
	BranchID        sql.NullString `json:"branch_id" db:"branch_id"`
	MerchantID      sql.NullString `json:"merchant_id" db:"merchant_id"`
	IsLocked        bool           `json:"is_locked" db:"is_locked"`
	IsFirstLogin    bool           `json:"is_first_login" db:"is_first_login"`
	LoggingAttempts int            `json:"logging_attempts" db:"logging_attempts"`
}

func (u *User) Table() string {
	return "users"
}

func (u *User) Columns() []string {
	return []string{"id", "full_name", "phone_number", "password", "passcode_hash", "role", "branch_id", "merchant_id", "is_locked", "is_first_login", "logging_attempts", "deleted_at", "is_deleted"}
}

func (u *User) Values() []any {
	return []any{u.ID, u.FullName, u.PhoneNumber, u.Password, u.PasscodeHash, u.Role, u.BranchID, u.MerchantID, u.IsLocked, u.IsFirstLogin, u.LoggingAttempts, u.DeletedAt, u.IsDeleted}
}

func (u *User) Addr() []any {
	return []any{&u.ID, &u.FullName, &u.PhoneNumber, &u.Password, &u.PasscodeHash, &u.Role, &u.BranchID, &u.MerchantID, &u.IsLocked, &u.IsFirstLogin, &u.LoggingAttempts, &u.DeletedAt, &u.IsDeleted, &u.CreatedAt, &u.UpdatedAt}
}

func (u *User) ToDTO() UserDTO {
	dto := UserDTO{
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
	if u.MerchantID.Valid {
		dto.MerchantID = u.MerchantID.String
	}
	return dto
}
