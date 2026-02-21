package users

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user User) error
	GetUserByID(ctx context.Context, id string) (User, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (User, error)
	UpdateUser(ctx context.Context, user User) error
	SetPassword(ctx context.Context, phoneNumber, id, password string, isFirstLogin bool) error
	DeleteUser(ctx context.Context, id string) error
	CheckUserExistsByPhoneNumber(ctx context.Context, phoneNumber string) error
	GetAllUsers(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*User], error)
	GetUserByBranchID(ctx context.Context, branchID string) (User, error)
	UpdateLoggingAttempts(ctx context.Context, id string, attempts int) error
	ResetLoggingAttempts(ctx context.Context, id string) error
	LockUser(ctx context.Context, id string) error
	UnDeleteUser(ctx context.Context, id string) error
}

type userRepository struct {
	dal     *common.DAL[*User]
	joinDAL *common.JoinDAL
	logger  config.Logger
}

func NewUserRepository(dal *common.DAL[*User], joinDAL *common.JoinDAL, logger config.Logger) UserRepository {
	return &userRepository{dal: dal, joinDAL: joinDAL, logger: logger}
}

func (r *userRepository) CreateUser(ctx context.Context, user User) error {
	_, err := r.dal.Create(ctx, &user)
	if err != nil {
		r.logger.Error("failed to create user", "error", err)
		return err
	}
	return nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (User, error) {
	filter := map[string]any{"id": id, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return User{}, common.ErrUserNotFound
		}
		r.logger.Error("failed to get user by ID", "error", err)
		return User{}, err
	}
	return *result, nil
}

func (r *userRepository) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (User, error) {
	filter := map[string]any{"phone_number": phoneNumber, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return User{}, common.ErrUserNotFound
		}
		r.logger.Error("failed to get user by phone number", "error", err)
		return User{}, err
	}
	return *result, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user User) error {
	r.logger.Info("Updating user and ID is", "user", user, "id", user.ID)

	filter := map[string]any{"id": user.ID}
	updates := map[string]any{
		"full_name":        user.FullName,
		"phone_number":     user.PhoneNumber,
		"password":         user.Password,
		"role":             user.Role,
		"branch_id":        user.BranchID,
		"is_locked":        user.IsLocked,
		"is_first_login":   user.IsFirstLogin,
		"logging_attempts": user.LoggingAttempts,
		"deleted_at":       user.DeletedAt,
		"is_deleted":       user.IsDeleted,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return common.ErrUserNotFound
		}
		r.logger.Error("failed to update user", "error", err)
		return err
	}
	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	err := r.dal.Delete(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return common.ErrUserNotFound
		}
		r.logger.Error("failed to delete user", "error", err)
		return err
	}
	return nil
}

func (r *userRepository) SetPassword(ctx context.Context, phoneNumber, id, password string, isFirstLogin bool) error {
	updates := map[string]any{
		"password": password,
	}
	if isFirstLogin {
		updates["is_first_login"] = false
	}

	filter := map[string]any{"id": id}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return common.ErrUserNotFound
		}
		r.logger.Error("failed to set password", "error", err)
		return err
	}
	return nil
}

func (r *userRepository) CheckUserExistsByPhoneNumber(ctx context.Context, phoneNumber string) error {
	filter := map[string]any{"phone_number": phoneNumber, "is_deleted": false}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		r.logger.Error("failed to check user exists by phone number", "error", err)
		return err
	}
	return common.ErrUserWithInformationAlreadyExists
}

func (r *userRepository) GetAllUsers(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*User], error) {
	role, _ := middleware.GetRoleFromContext(ctx)
	filters := map[string]any{}
	if filter.Search != "" {
		filters["full_name"] = common.ILike(filter.Search)
	}

	merchantID, hasMerchantID := filter.Filter["merchant_id"].(string)
	if hasMerchantID && merchantID != "" {
		// Users don't have merchant_id; filter via join: user -> branch -> merchant
		return r.listUsersByMerchant(ctx, filter, merchantID, role)
	}

	if len(filter.Filter) > 0 {
		if roleVal, ok := filter.Filter["role"].(string); ok && roleVal != "" {
			filters["role"] = roleVal
		}
		if branchID, ok := filter.Filter["branch_id"].(string); ok && branchID != "" {
			filters["branch_id"] = branchID
		}
	}

	offset := (filter.Page - 1) * filter.Limit
	var results []*User
	var err error
	if role == string(RoleAdmin) {
		results, err = r.dal.ListIncludeDeleted(ctx, filters, filter.Limit, offset)
	} else {
		results, err = r.dal.List(ctx, filters, filter.Page, filter.Limit)
	}
	if err != nil {
		r.logger.Error("failed to get all users", "error", err)
		return nil, err
	}
	return &common.PaginatedResponse[[]*User]{
		Data: results,
		Meta: common.BuildPaginationMeta(int64(len(results)), filter.Page, filter.Limit),
	}, nil
}

// listUsersByMerchant returns users whose branch belongs to the given merchant.
func (r *userRepository) listUsersByMerchant(ctx context.Context, filter common.Filter, merchantID string, role string) (*common.PaginatedResponse[[]*User], error) {
	offset := (filter.Page - 1) * filter.Limit

	// JOIN users with branches to filter by merchant_id
	// Include users with branch_id IN (branches of merchant) OR super_admin (branch_id may be null)
	baseQuery := `
		SELECT u.id, u.full_name, u.phone_number, u.password, u.role, u.branch_id, u.merchant_id, u.is_locked, u.is_first_login, u.logging_attempts, u.deleted_at, u.is_deleted, u.created_at, u.updated_at
		FROM users u
		INNER JOIN branches b ON u.branch_id = b.id AND b.merchant_id = $1
		WHERE 1=1`
	args := []any{merchantID}
	argNum := 2

	var conditions []string
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(" AND u.full_name ILIKE $%d", argNum))
		args = append(args, common.ILikePattern(filter.Search))
		argNum++
	}
	if roleVal, ok := filter.Filter["role"].(string); ok && roleVal != "" {
		conditions = append(conditions, fmt.Sprintf(" AND u.role = $%d", argNum))
		args = append(args, roleVal)
		argNum++
	}
	if role != string(RoleAdmin) {
		conditions = append(conditions, " AND u.is_deleted = FALSE")
	}

	query := baseQuery + strings.Join(conditions, "") + " ORDER BY u.created_at DESC LIMIT $" + fmt.Sprint(argNum) + " OFFSET $" + fmt.Sprint(argNum+1)
	args = append(args, filter.Limit, offset)

	results, err := common.QueryRows(r.joinDAL, ctx, query, args, func(rows *sql.Rows) (*User, error) {
		item := &User{}
		if err := rows.Scan(item.Addr()...); err != nil {
			return nil, err
		}
		return item, nil
	})
	if err != nil {
		r.logger.Error("failed to list users by merchant", "error", err)
		return nil, err
	}
	return &common.PaginatedResponse[[]*User]{
		Data: results,
		Meta: common.BuildPaginationMeta(int64(len(results)), filter.Page, filter.Limit),
	}, nil
}

func (r *userRepository) GetUserByBranchID(ctx context.Context, branchID string) (User, error) {
	filter := map[string]any{"branch_id": branchID, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return User{}, common.ErrUserNotFound
		}
		r.logger.Error("failed to get user by branch ID", "error", err)
		return User{}, err
	}
	return *result, nil
}

func (r *userRepository) UpdateLoggingAttempts(ctx context.Context, id string, attempts int) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{"logging_attempts": attempts}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to update logging attempts", "error", err)
		return err
	}
	return nil
}

func (r *userRepository) ResetLoggingAttempts(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{"logging_attempts": 0}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to reset logging attempts", "error", err)
		return err
	}
	return nil
}

func (r *userRepository) LockUser(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{"is_locked": true}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to lock user", "error", err)
		return err
	}
	return nil
}

func (r *userRepository) UnDeleteUser(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{"is_deleted": false}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to undelete user", "error", err)
		return err
	}
	return nil
}
