package merchant

import (
	"context"
	"database/sql"
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"lazeez-core/internal/users"
)

type MerchantRepository interface {
	Create(ctx context.Context, merchant Merchant) (*MerchantDTO, error)
	Get(ctx context.Context, id string) (*MerchantDTO, error)
	Update(ctx context.Context, merchant Merchant) error
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context) ([]*MerchantDTO, error)
	CheckExists(ctx context.Context, name string) error
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
	m.id, m.name, m.logo, m.created_at, m.updated_at, m.deleted_at, m.is_deleted,
	COALESCE((
		SELECT json_agg(json_build_object(
			'id', b.id, 'merchant_id', b.merchant_id, 'branch_name', b.branch_name,
			'address', b.address, 'phone_number', b.phone_number,
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
	m.id, m.name, m.logo, m.created_at, m.updated_at, m.deleted_at, m.is_deleted,
	COALESCE((
		SELECT json_agg(json_build_object(
			'id', b.id, 'merchant_id', b.merchant_id, 'branch_name', b.branch_name,
			'address', b.address, 'phone_number', b.phone_number,
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
		if err == sql.ErrNoRows {
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
		"name":       merchant.Name,
		"logo":       merchant.Logo,
		"deleted_at": merchant.DeletedAt,
		"is_deleted": merchant.IsDeleted,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if err == sql.ErrNoRows {
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

func (r *merchantRepository) GetAll(ctx context.Context) ([]*MerchantDTO, error) {
	role, _ := middleware.GetRoleFromContext(ctx)

	var baseQuery string
	if role == string(users.RoleAdmin) {
		// Super admin sees deleted and non-deleted merchants, branches, and users
		baseQuery = merchantWithRelationsAll
	} else {
		// Others only see non-deleted merchants, branches, and users
		baseQuery = merchantWithRelationsActive
	}
	query := baseQuery + " ORDER BY m.created_at DESC"

	results, err := common.QueryRows(r.joinDAL, ctx, query, nil, func(rows *sql.Rows) (*MerchantDTO, error) {
		var dto MerchantDTO
		if err := r.scanMerchantWithRelationsFromRows(rows, &dto); err != nil {
			return nil, err
		}
		return &dto, nil
	})
	if err != nil {
		r.logger.Error("failed to get all merchants", "error", err)
		return nil, err
	}
	return results, nil
}

// scanMerchantWithRelations scans a merchant row with JSON branches and users into MerchantDTO.
func (r *merchantRepository) scanMerchantWithRelations(row *sql.Row, dto *MerchantDTO) error {
	var branchesJSON, usersJSON []byte
	var deletedAt sql.NullTime
	err := row.Scan(
		&dto.ID, &dto.Name, &dto.Logo,
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
		&dto.ID, &dto.Name, &dto.Logo,
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

func (r *merchantRepository) CheckExists(ctx context.Context, name string) error {
	filter := map[string]any{"name": name, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		r.logger.Error("failed to check if merchant exists", "error", err)
		return err
	}

	// If merchant exists, return error
	if result.ID != "" {
		return common.ErrMerchantAlreadyExists
	}
	return nil
}
