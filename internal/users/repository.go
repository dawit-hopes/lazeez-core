package users

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"

	"github.com/lib/pq"
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
	GetWaitersByBranch(ctx context.Context, branchID string) ([]User, error)
	UpdateLoggingAttempts(ctx context.Context, id string, attempts int) error
	ResetLoggingAttempts(ctx context.Context, id string) error
	LockUser(ctx context.Context, id string) error
	UnDeleteUser(ctx context.Context, id string) error
	ResolveMerchantContext(ctx context.Context, branchID, storedMerchantID string) (merchantID, branchType string, err error)
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
	common.NormalizeFilter(&filter)
	roleStr, ok := middleware.GetRoleFromContext(ctx)
	if !ok {
		return nil, common.ErrUnAuthorized
	}
	viewerRole := Role(roleStr)

	merchantID, _ := middleware.GetMerchantIDFromContext(ctx)
	branchID, _ := middleware.GetBranchIDFromContext(ctx)

	var branchType BranchType
	if viewerRole == RoleBranchManager {
		_, bt, err := r.ResolveMerchantContext(ctx, branchID, merchantID)
		if err != nil {
			return nil, err
		}
		branchType = BranchType(bt)
	}

	viewable := ViewableRoles(viewerRole, branchType)
	if len(viewable) == 0 {
		return &common.PaginatedResponse[[]*User]{
			Data: []*User{},
			Meta: common.BuildPaginationMeta(0, filter.Page, filter.Limit),
		}, nil
	}

	// Optional role filter must stay within viewable roles.
	if roleVal, ok := filter.Filter["role"].(string); ok && roleVal != "" {
		if !CanViewerSeeUser(viewerRole, Role(roleVal), branchType) {
			return &common.PaginatedResponse[[]*User]{
				Data: []*User{},
				Meta: common.BuildPaginationMeta(0, filter.Page, filter.Limit),
			}, nil
		}
		viewable = []Role{Role(roleVal)}
	}

	return r.listUsersScoped(ctx, filter, viewerRole, viewable, merchantID, branchID)
}

func (r *userRepository) listUsersScoped(ctx context.Context, filter common.Filter, viewerRole Role, viewable []Role, merchantID, branchID string) (*common.PaginatedResponse[[]*User], error) {
	offset := (filter.Page - 1) * filter.Limit
	roleStrings := rolesToStrings(viewable)

	baseSelect := `
		SELECT u.id, u.full_name, u.phone_number, u.password, u.passcode_hash, u.role, u.branch_id, u.merchant_id,
			u.is_locked, u.is_first_login, u.logging_attempts, u.deleted_at, u.is_deleted, u.created_at, u.updated_at
		FROM users u`

	var conditions []string
	var args []any
	argNum := 1

	nextArg := func(v any) string {
		placeholder := fmt.Sprintf("$%d", argNum)
		args = append(args, v)
		argNum++
		return placeholder
	}

	conditions = append(conditions, fmt.Sprintf("u.role = ANY(%s)", nextArg(pq.Array(roleStrings))))

	switch viewerRole {
	case RoleAdmin:
		if filterMerchantID, ok := filter.Filter["merchant_id"].(string); ok && filterMerchantID != "" {
			conditions = append(conditions, fmt.Sprintf(`(
				u.merchant_id = %s OR EXISTS (
					SELECT 1 FROM branches b WHERE b.id = u.branch_id AND b.merchant_id = %s
				)
			)`, nextArg(filterMerchantID), nextArg(filterMerchantID)))
		}
	case RoleSuperBranchManager:
		if merchantID == "" {
			return nil, common.ErrUnAuthorized
		}
		conditions = append(conditions, fmt.Sprintf(`(
			u.merchant_id = %s OR EXISTS (
				SELECT 1 FROM branches b WHERE b.id = u.branch_id AND b.merchant_id = %s AND b.is_deleted = FALSE
			)
		)`, nextArg(merchantID), nextArg(merchantID)))
		if filterBranchID, ok := filter.Filter["branch_id"].(string); ok && filterBranchID != "" {
			conditions = append(conditions, fmt.Sprintf(`EXISTS (
				SELECT 1 FROM branches b WHERE b.id = %s AND b.merchant_id = %s AND b.is_deleted = FALSE
			)`, nextArg(filterBranchID), nextArg(merchantID)))
			conditions = append(conditions, fmt.Sprintf("u.branch_id = %s", nextArg(filterBranchID)))
		}
		conditions = append(conditions, "u.is_deleted = FALSE")
	case RoleBranchManager:
		if branchID == "" {
			return nil, common.ErrUnAuthorized
		}
		conditions = append(conditions, fmt.Sprintf("u.branch_id = %s", nextArg(branchID)))
		conditions = append(conditions, "u.is_deleted = FALSE")
	default:
		return nil, common.ErrUnAuthorized
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("u.full_name ILIKE %s", nextArg(common.ILikePattern(filter.Search))))
	}

	where := " WHERE " + strings.Join(conditions, " AND ")
	limitPH := nextArg(filter.Limit)
	offsetPH := nextArg(offset)
	query := baseSelect + where + " ORDER BY u.created_at DESC LIMIT " + limitPH + " OFFSET " + offsetPH

	results, err := common.QueryRows(r.joinDAL, ctx, query, args, func(rows *sql.Rows) (*User, error) {
		item := &User{}
		if err := rows.Scan(item.Addr()...); err != nil {
			return nil, err
		}
		return item, nil
	})
	if err != nil {
		r.logger.Error("failed to list scoped users", "error", err)
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

// GetWaitersByBranch returns non-deleted waiter accounts in a branch that have a PIN set.
// Used to resolve a submitted PIN to a specific waiter for order attribution.
func (r *userRepository) GetWaitersByBranch(ctx context.Context, branchID string) ([]User, error) {
	filters := map[string]any{
		"branch_id":  branchID,
		"role":       string(RoleWaiter),
		"is_deleted": false,
	}
	results, err := r.dal.List(ctx, filters, 1, 1000)
	if err != nil {
		r.logger.Error("failed to list waiters by branch", "error", err)
		return nil, err
	}
	waiters := make([]User, 0, len(results))
	for _, u := range results {
		if u != nil {
			waiters = append(waiters, *u)
		}
	}
	return waiters, nil
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

func (r *userRepository) ResolveMerchantContext(ctx context.Context, branchID, storedMerchantID string) (string, string, error) {
	if storedMerchantID != "" {
		const query = `SELECT branch_type FROM merchants WHERE id = $1 AND is_deleted = FALSE LIMIT 1`
		var branchType string
		err := r.joinDAL.QueryRow(ctx, query, []any{storedMerchantID}, func(row *sql.Row) error {
			return row.Scan(&branchType)
		})
		if err != nil {
			if err == sql.ErrNoRows {
				return "", "", common.ErrMerchantNotFound
			}
			r.logger.Error("failed to resolve merchant context", "merchant_id", storedMerchantID, "error", err)
			return "", "", err
		}
		return storedMerchantID, branchType, nil
	}

	if branchID != "" {
		const query = `
			SELECT b.merchant_id, m.branch_type
			FROM branches b
			INNER JOIN merchants m ON m.id = b.merchant_id AND m.is_deleted = FALSE
			WHERE b.id = $1 AND b.is_deleted = FALSE
			LIMIT 1`
		var merchantID, branchType string
		err := r.joinDAL.QueryRow(ctx, query, []any{branchID}, func(row *sql.Row) error {
			return row.Scan(&merchantID, &branchType)
		})
		if err != nil {
			if err == sql.ErrNoRows {
				return "", "", common.ErrBranchNotFound
			}
			r.logger.Error("failed to resolve merchant context from branch", "branch_id", branchID, "error", err)
			return "", "", err
		}
		return merchantID, branchType, nil
	}

	return "", "", nil
}
