package merchant

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"lazeez-core/internal/users"
	"strings"
)

type MerchantRepository interface {
	Create(ctx context.Context, merchant Merchant) (*MerchantDTO, error)
	Get(ctx context.Context, id string) (*MerchantDTO, error)
	Update(ctx context.Context, merchant Merchant) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
	GetAll(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*MerchantDTO], error)
	CheckExists(ctx context.Context, name string, excludeMerchantID string) error
}

type merchantRepository struct {
	dal     *common.DAL[*Merchant]
	joinDAL *common.JoinDAL
	logger  config.Logger
}

func NewMerchantRepository(
	dal *common.DAL[*Merchant],
	joinDAL *common.JoinDAL,
	logger config.Logger,
) MerchantRepository {
	return &merchantRepository{
		dal:     dal,
		joinDAL: joinDAL,
		logger:  logger,
	}
}

// merchantWithRelationsActive selects merchants and their related branches and
// users, filtering out soft-deleted records everywhere.
const merchantWithRelationsActive = `
SELECT 
	m.id, m.name, m.branch_type, m.logo, m.created_at, m.updated_at, m.deleted_at, m.is_deleted,
	COALESCE((
		SELECT json_agg(json_build_object(
			'id', b.id, 'merchant_id', b.merchant_id, 'branch_name', b.branch_name,
			'address', b.address, 'phone_number', b.phone_number, 'branch_type', m.branch_type,
			'created_at', b.created_at, 'updated_at', b.updated_at, 'deleted_at', b.deleted_at, 'is_deleted', b.is_deleted
		))
		FROM branches b
		WHERE b.merchant_id = m.id AND b.is_deleted = FALSE
	), '[]'::json)::text AS branches,
	COALESCE((
		SELECT json_agg(json_build_object(
			'id', u.id, 'phone_number', u.phone_number, 'full_name', u.full_name,
			'role', u.role, 'branch_id', COALESCE(u.branch_id::text, ''),
			'is_locked', u.is_locked, 'is_first_login', u.is_first_login, 'logging_attempts', u.logging_attempts,
			'created_at', u.created_at, 'updated_at', u.updated_at, 'deleted_at', u.deleted_at, 'is_deleted', u.is_deleted
		))
		FROM users u
		INNER JOIN branches b ON u.branch_id = b.id
		WHERE b.merchant_id = m.id AND u.is_deleted = FALSE AND b.is_deleted = FALSE
	), '[]'::json)::text AS users
FROM merchants m
WHERE m.is_deleted = FALSE
`

// merchantWithRelationsAll selects merchants and their related branches and
// users without filtering out soft-deleted records. Intended for super_admin.
const merchantWithRelationsAll = `
SELECT 
	m.id, m.name, m.branch_type, m.logo, m.created_at, m.updated_at, m.deleted_at, m.is_deleted,
	COALESCE((
		SELECT json_agg(json_build_object(
			'id', b.id, 'merchant_id', b.merchant_id, 'branch_name', b.branch_name,
			'address', b.address, 'phone_number', b.phone_number, 'branch_type', m.branch_type,
			'created_at', b.created_at, 'updated_at', b.updated_at, 'deleted_at', b.deleted_at, 'is_deleted', b.is_deleted
		))
		FROM branches b
		WHERE b.merchant_id = m.id
	), '[]'::json)::text AS branches,
	COALESCE((
		SELECT json_agg(json_build_object(
			'id', u.id, 'phone_number', u.phone_number, 'full_name', u.full_name,
			'role', u.role, 'branch_id', COALESCE(u.branch_id::text, ''),
			'is_locked', u.is_locked, 'is_first_login', u.is_first_login, 'logging_attempts', u.logging_attempts,
			'created_at', u.created_at, 'updated_at', u.updated_at, 'deleted_at', u.deleted_at, 'is_deleted', u.is_deleted
		))
		FROM users u
		INNER JOIN branches b ON u.branch_id = b.id
		WHERE b.merchant_id = m.id
	), '[]'::json)::text AS users
FROM merchants m
`

// merchantListActive lists merchants with branch/user counts only (no nested JSON).
const merchantListActive = `
SELECT 
	m.id, m.name, m.branch_type, m.logo, m.created_at, m.updated_at, m.deleted_at, m.is_deleted,
	(SELECT COUNT(*)::int FROM branches b WHERE b.merchant_id = m.id AND b.is_deleted = FALSE) AS total_branches,
	(SELECT COUNT(DISTINCT u.id)::int FROM users u
	 WHERE u.is_deleted = FALSE
	 AND (u.merchant_id = m.id OR EXISTS (
	   SELECT 1 FROM branches b WHERE b.id = u.branch_id AND b.merchant_id = m.id AND b.is_deleted = FALSE
	 ))) AS total_users
FROM merchants m
WHERE m.is_deleted = FALSE
`

// merchantListAll lists merchants with counts, including soft-deleted relations. Intended for super_admin.
const merchantListAll = `
SELECT 
	m.id, m.name, m.branch_type, m.logo, m.created_at, m.updated_at, m.deleted_at, m.is_deleted,
	(SELECT COUNT(*)::int FROM branches b WHERE b.merchant_id = m.id) AS total_branches,
	(SELECT COUNT(DISTINCT u.id)::int FROM users u
	 WHERE u.merchant_id = m.id OR EXISTS (
	   SELECT 1 FROM branches b WHERE b.id = u.branch_id AND b.merchant_id = m.id
	 )) AS total_users
FROM merchants m
`

func (r *merchantRepository) Create(ctx context.Context, merchant Merchant) (*MerchantDTO, error) {
	_, err := r.dal.Create(ctx, &merchant)
	if err != nil {
		r.logger.Error("failed to create merchant", "error", err)
		return nil, err
	}

	merchantDTO := merchant.ToDTO()
	return &merchantDTO, nil
}

func (r *merchantRepository) Get(ctx context.Context, id string) (*MerchantDTO, error) {
	// For single get we still filter out deleted merchants and related data.
	query := merchantWithRelationsActive + " AND m.id = $1"
	var dto MerchantDTO
	err := r.joinDAL.QueryRow(ctx, query, []any{id}, func(row *sql.Row) error {
		return r.scanMerchantWithRelations(row, &dto)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("merchant not found", "error", err)
			return nil, common.ErrMerchantNotFound
		}
		r.logger.Error("failed to get merchant", "error", err)
		return nil, err
	}
	return &dto, nil
}

func (r *merchantRepository) Update(ctx context.Context, merchant Merchant) error {
	filter := map[string]any{"id": merchant.ID}
	updates := map[string]any{
		"name":        merchant.Name,
		"branch_type": merchant.BranchType,
		"logo":        merchant.Logo,
		"deleted_at":  merchant.DeletedAt,
		"is_deleted":  merchant.IsDeleted,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("merchant not found", "error", err)
			return common.ErrMerchantNotFound
		}
		r.logger.Error("failed to update merchant", "error", err)
		return err
	}
	return nil
}

func (r *merchantRepository) Delete(ctx context.Context, id string) error {
	const deleteMerchantCascadeQuery = `
WITH updated_users AS (
	UPDATE users u
	SET is_deleted = TRUE, updated_at = NOW(), deleted_at = NOW()
	WHERE u.branch_id IN (
		SELECT b.id FROM branches b
		WHERE b.merchant_id = $1 AND b.is_deleted = FALSE
	) AND u.is_deleted = FALSE
),
updated_branches AS (
	UPDATE branches b
	SET is_deleted = TRUE, updated_at = NOW(), deleted_at = NOW()
	WHERE b.merchant_id = $1 AND b.is_deleted = FALSE
)
UPDATE merchants m
SET is_deleted = TRUE, updated_at = NOW(), deleted_at = NOW()
WHERE m.id = $1 AND m.is_deleted = FALSE;
`

	result, err := r.joinDAL.Exec(ctx, deleteMerchantCascadeQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("merchant not found", "error", err)
			return common.ErrMerchantNotFound
		}
		r.logger.Error("failed to cascade delete merchant", "merchant_id", id, "error", err)
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		r.logger.Error("merchant not found", "merchant_id", id)
		return common.ErrMerchantNotFound
	}

	return nil
}

func (r *merchantRepository) GetAll(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*MerchantDTO], error) {
	common.NormalizeFilter(&filter)
	role, _ := middleware.GetRoleFromContext(ctx)
	isSuperAdmin := role == string(users.RoleAdmin)

	baseQuery := merchantListActive
	if isSuperAdmin {
		baseQuery = merchantListAll
	}

	var conditions []string
	var args []any
	argIndex := 0
	nextArg := func(val any) string {
		argIndex++
		args = append(args, val)
		return fmt.Sprintf("$%d", argIndex)
	}

	if filter.Search != "" {
		searchPattern := common.ILikePattern(filter.Search)
		p := nextArg(searchPattern)
		var searchByBranch, searchByUser string
		if isSuperAdmin {
			searchByBranch = fmt.Sprintf("EXISTS (SELECT 1 FROM branches b WHERE b.merchant_id = m.id AND (b.branch_name ILIKE %s OR b.address ILIKE %s OR b.phone_number ILIKE %s))", p, p, p)
			searchByUser = fmt.Sprintf("EXISTS (SELECT 1 FROM users u INNER JOIN branches b ON u.branch_id = b.id WHERE b.merchant_id = m.id AND (u.full_name ILIKE %s OR u.phone_number ILIKE %s))", p, p)
		} else {
			searchByBranch = fmt.Sprintf("EXISTS (SELECT 1 FROM branches b WHERE b.merchant_id = m.id AND b.is_deleted = FALSE AND (b.branch_name ILIKE %s OR b.address ILIKE %s OR b.phone_number ILIKE %s))", p, p, p)
			searchByUser = fmt.Sprintf("EXISTS (SELECT 1 FROM users u INNER JOIN branches b ON u.branch_id = b.id WHERE b.merchant_id = m.id AND u.is_deleted = FALSE AND b.is_deleted = FALSE AND (u.full_name ILIKE %s OR u.phone_number ILIKE %s))", p, p)
		}
		conditions = append(conditions, fmt.Sprintf("(m.name ILIKE %s OR %s OR %s)", p, searchByBranch, searchByUser))
	}

	if filter.Filter != nil {
		if branchType, ok := filter.Filter["branch_type"].(string); ok && branchType != "" {
			conditions = append(conditions, fmt.Sprintf("m.branch_type = %s", nextArg(branchType)))
		}
	}

	query := baseQuery
	if len(conditions) > 0 {
		joiner := " AND "
		if isSuperAdmin {
			query += " WHERE " + strings.Join(conditions, joiner)
		} else {
			query += joiner + strings.Join(conditions, joiner)
		}
	}

	offset := (filter.Page - 1) * filter.Limit
	query += fmt.Sprintf(" ORDER BY m.created_at DESC LIMIT %s OFFSET %s", nextArg(filter.Limit), nextArg(offset))

	results, err := common.QueryRows(r.joinDAL, ctx, query, args, func(rows *sql.Rows) (*MerchantDTO, error) {
		var dto MerchantDTO
		if err := r.scanMerchantListFromRows(rows, &dto); err != nil {
			return nil, err
		}
		return &dto, nil
	})
	if err != nil {
		r.logger.Error("failed to get all merchants", "error", err)
		return nil, err
	}
	return &common.PaginatedResponse[[]*MerchantDTO]{
		Data: results,
		Meta: common.BuildPaginationMeta(int64(len(results)), filter.Page, filter.Limit),
	}, nil
}

// scanMerchantListFromRows scans a merchant list row with aggregate counts only.
func (r *merchantRepository) scanMerchantListFromRows(rows *sql.Rows, dto *MerchantDTO) error {
	var deletedAt sql.NullTime
	err := rows.Scan(
		&dto.ID, &dto.Name, &dto.BranchType, &dto.Logo,
		&dto.CreatedAt, &dto.UpdatedAt, &deletedAt, &dto.IsDeleted,
		&dto.TotalBranches, &dto.TotalUsers,
	)
	if err != nil {
		return err
	}
	dto.DeletedAt = common.ToNullTimePtr(deletedAt)
	return nil
}

// scanMerchantWithRelations scans a merchant row with JSON branches and users into MerchantDTO.
func (r *merchantRepository) scanMerchantWithRelations(row *sql.Row, dto *MerchantDTO) error {
	var branchesJSON, usersJSON []byte
	var deletedAt sql.NullTime
	err := row.Scan(
		&dto.ID, &dto.Name, &dto.BranchType, &dto.Logo,
		&dto.CreatedAt, &dto.UpdatedAt, &deletedAt, &dto.IsDeleted,
		&branchesJSON, &usersJSON,
	)
	if err != nil {
		return err
	}
	dto.DeletedAt = common.ToNullTimePtr(deletedAt)
	return r.unmarshalRelations(dto, branchesJSON, usersJSON)
}

func (r *merchantRepository) scanMerchantWithRelationsFromRows(rows *sql.Rows, dto *MerchantDTO) error {
	var branchesJSON, usersJSON []byte
	var deletedAt sql.NullTime
	err := rows.Scan(
		&dto.ID, &dto.Name, &dto.BranchType, &dto.Logo,
		&dto.CreatedAt, &dto.UpdatedAt, &deletedAt, &dto.IsDeleted,
		&branchesJSON, &usersJSON,
	)
	if err != nil {
		return err
	}
	dto.DeletedAt = common.ToNullTimePtr(deletedAt)
	return r.unmarshalRelations(dto, branchesJSON, usersJSON)
}

func (r *merchantRepository) unmarshalRelations(dto *MerchantDTO, branchesJSON, usersJSON []byte) error {
	var branches []*branch.BranchResponse
	if err := json.Unmarshal(branchesJSON, &branches); err != nil {
		return err
	}
	if branches == nil {
		branches = []*branch.BranchResponse{}
	}

	var userDTOs []*users.UserDTO
	if err := json.Unmarshal(usersJSON, &userDTOs); err != nil {
		return err
	}
	if userDTOs == nil {
		userDTOs = []*users.UserDTO{}
	}

	dto.Branches = branches
	dto.Users = userDTOs
	dto.TotalBranches = len(branches)
	dto.TotalUsers = len(userDTOs)
	return nil
}

func (r *merchantRepository) CheckExists(ctx context.Context, name string, excludeMerchantID string) error {
	query := `
		SELECT id FROM merchants
		WHERE LOWER(name) = LOWER($1) AND is_deleted = FALSE`
	args := []any{name}
	if excludeMerchantID != "" {
		query += ` AND id != $2`
		args = append(args, excludeMerchantID)
	}
	query += ` LIMIT 1`

	var id string
	err := r.joinDAL.QueryRow(ctx, query, args, func(row *sql.Row) error {
		return row.Scan(&id)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check if merchant exists", "error", err)
		return err
	}

	return common.ErrMerchantAlreadyExists
}

func (r *merchantRepository) UnDelete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": false,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("merchant not found", "error", err)
			return common.ErrMerchantNotFound
		}
		r.logger.Error("failed to undelete merchant", "error", err)
		return err
	}
	return nil
}
