package menu

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"lazeez-core/config"
	"lazeez-core/internal/category"
	"lazeez-core/internal/common"
	"lazeez-core/internal/ingredient"
)

type ListScope string

const (
	ScopeMaster          ListScope = "master"
	ScopeBranchEffective ListScope = "branch"
	ScopeBranchManage    ListScope = "branch_manage"
	ScopeAllBranches     ListScope = "all_branches"
)

type MenuRepository interface {
	Create(ctx context.Context, menu Menu) error
	Get(ctx context.Context, id string, branchID string) (Menu, error)
	GetMaster(ctx context.Context, id string, merchantID string) (Menu, error)
	Update(ctx context.Context, menu Menu) error
	Delete(ctx context.Context, id string, branchID string) error
	UnDelete(ctx context.Context, id string, branchID string) error
	List(ctx context.Context, filter common.Filter, branchID string) (*common.PaginatedResponse[[]*Menu], error)
	ListScoped(ctx context.Context, filter common.Filter, scope ListScope, branchID, merchantID string) (*common.PaginatedResponse[[]*Menu], error)
	CheckExists(ctx context.Context, name, branchID string) error
	CheckMasterExists(ctx context.Context, name, merchantID string) error
	SetBranchOverride(ctx context.Context, branchID, menuID string, isAvailable bool) error
	SetBranchExcluded(ctx context.Context, branchID, menuID string, excluded bool) error
	RemoveBranchOverride(ctx context.Context, branchID, menuID string) error
	AssignOrphanMasterMenus(ctx context.Context, merchantID string) error
	GetBranchMenuSnapshots(ctx context.Context, branchID string, menuIDs []string) (map[string]BranchMenuSnapshot, error)

	// public repository
	ListMenus(ctx context.Context, filter common.Filter, reference string) (*PublicMenuCatalogResponse, error)
}

type menuRepository struct {
	dal    *common.DAL[*Menu]
	join   *common.JoinDAL
	logger config.Logger
}

func NewMenuRepository(dal *common.DAL[*Menu], join *common.JoinDAL, logger config.Logger) MenuRepository {
	return &menuRepository{dal: dal, join: join, logger: logger}
}

func (r *menuRepository) Create(ctx context.Context, menu Menu) error {
	_, err := r.dal.Create(ctx, &menu)
	if err != nil {
		r.logger.Error("failed to create menu", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *menuRepository) Get(ctx context.Context, id string, branchID string) (Menu, error) {
	if branchID != "" {
		if menu, err := r.getBranchOwned(ctx, id, branchID); err == nil {
			return menu, nil
		} else if !errors.Is(err, common.ErrMenuNotFound) {
			return Menu{}, err
		}
		return r.getMasterForBranch(ctx, id, branchID)
	}
	return Menu{}, common.ErrMenuNotFound
}

func (r *menuRepository) getBranchOwned(ctx context.Context, id, branchID string) (Menu, error) {
	filter := map[string]any{"id": id, "branch_id": branchID}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Menu{}, common.ErrMenuNotFound
		}
		r.logger.Error("failed to get branch menu", "error", err)
		return Menu{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *menuRepository) getMasterForBranch(ctx context.Context, id, branchID string) (Menu, error) {
	query := `
		SELECT m.id, m.name, m.image, m.deleted_at, m.is_deleted, m.branch_id, m.merchant_id,
			m.is_fasting,
			COALESCE(o.is_available, m.is_available) AS is_available,
			m.description, m.price, m.ingredients, m.category_id, m.modifiers, m.preparation_time,
			m.created_at, m.updated_at,
			COALESCE(o.is_excluded, FALSE) AS is_excluded
		FROM menus m
		INNER JOIN branches b ON b.id = $2 AND b.is_deleted = FALSE
		LEFT JOIN branch_menu_overrides o
			ON o.menu_id = m.id AND o.branch_id = $2 AND o.is_deleted = FALSE
		WHERE m.id = $1 AND m.is_deleted = FALSE AND m.branch_id IS NULL
			AND m.merchant_id = b.merchant_id`

	var menu Menu
	var isExcluded bool
	err := r.join.QueryRow(ctx, query, []any{id, branchID}, func(row *sql.Row) error {
		return row.Scan(
			&menu.ID, &menu.Name, &menu.Image, &menu.DeletedAt, &menu.IsDeleted,
			&menu.BranchID, &menu.MerchantID, &menu.IsFasting, &menu.IsAvailable,
			&menu.Description, &menu.Price, &menu.Ingredients, &menu.CategoryID,
			&menu.Modifiers, &menu.PreparationTime, &menu.CreatedAt, &menu.UpdatedAt,
			&isExcluded,
		)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Menu{}, common.ErrMenuNotFound
		}
		r.logger.Error("failed to get master menu for branch", "error", err)
		return Menu{}, common.ErrInternalServerError
	}
	menu.Excluded = isExcluded
	menu.MasterItem = true
	if isExcluded {
		return Menu{}, common.ErrMenuNotFound
	}
	return menu, nil
}

func (r *menuRepository) GetMaster(ctx context.Context, id string, merchantID string) (Menu, error) {
	filter := map[string]any{"id": id, "branch_id": common.IsNull{}}
	if merchantID != "" {
		filter["merchant_id"] = merchantID
	}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Menu{}, common.ErrMenuNotFound
		}
		r.logger.Error("failed to get master menu", "error", err)
		return Menu{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *menuRepository) Update(ctx context.Context, menu Menu) error {
	filter := map[string]any{"id": menu.ID}
	updates := map[string]any{
		"name":             menu.Name,
		"image":            menu.Image,
		"description":      menu.Description,
		"price":            menu.Price,
		"ingredients":      menu.Ingredients,
		"category_id":      menu.CategoryID,
		"branch_id":        menu.BranchID,
		"merchant_id":      menu.MerchantID,
		"is_fasting":       menu.IsFasting,
		"is_available":     menu.IsAvailable,
		"deleted_at":       menu.DeletedAt,
		"is_deleted":       menu.IsDeleted,
		"preparation_time": menu.PreparationTime,
		"modifiers":        menu.Modifiers,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrMenuNotFound
		}
		r.logger.Error("failed to update menu", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *menuRepository) Delete(ctx context.Context, id string, branchID string) error {
	filter := map[string]any{"id": id}
	if branchID != "" {
		filter["branch_id"] = branchID
	} else {
		filter["branch_id"] = common.IsNull{}
	}
	updates := map[string]any{"is_deleted": true}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrMenuNotFound
		}
		r.logger.Error("failed to delete menu", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *menuRepository) UnDelete(ctx context.Context, id string, branchID string) error {
	filter := map[string]any{"id": id}
	if branchID != "" {
		filter["branch_id"] = branchID
	} else {
		filter["branch_id"] = common.IsNull{}
	}
	updates := map[string]any{"is_deleted": false}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrMenuNotFound
		}
		r.logger.Error("failed to undelete menu", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *menuRepository) List(ctx context.Context, filter common.Filter, branchID string) (*common.PaginatedResponse[[]*Menu], error) {
	scope := ScopeBranchEffective
	if branchID == "" {
		scope = ScopeMaster
	}
	return r.ListScoped(ctx, filter, scope, branchID, "")
}

func (r *menuRepository) ListScoped(ctx context.Context, filter common.Filter, scope ListScope, branchID, merchantID string) (*common.PaginatedResponse[[]*Menu], error) {
	common.NormalizeFilter(&filter)
	switch scope {
	case ScopeMaster:
		return r.listMaster(ctx, filter, merchantID)
	case ScopeAllBranches:
		return r.listAllBranchesEffective(ctx, filter, merchantID)
	case ScopeBranchManage:
		if branchID == "" {
			return nil, common.ErrUnAuthorized
		}
		return r.listBranchManage(ctx, filter, branchID)
	default:
		if branchID == "" {
			return nil, common.ErrUnAuthorized
		}
		return r.listBranchEffective(ctx, filter, branchID)
	}
}

func (r *menuRepository) listMaster(ctx context.Context, filter common.Filter, merchantID string) (*common.PaginatedResponse[[]*Menu], error) {
	if merchantID == "" {
		return nil, common.ErrUnAuthorized
	}
	filters := map[string]any{"branch_id": common.IsNull{}, "merchant_id": merchantID}
	if filter.Search != "" {
		filters["name"] = common.ILike(filter.Search)
	}
	page, limit := filter.PageLimit()
	menus, err := r.dal.List(ctx, filters, page, limit)
	if err != nil {
		r.logger.Error("failed to list master menus", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &common.PaginatedResponse[[]*Menu]{
		Data: menus,
		Meta: common.BuildPaginationMeta(int64(len(menus)), page, limit),
	}, nil
}

func (r *menuRepository) listBranchEffective(ctx context.Context, filter common.Filter, branchID string) (*common.PaginatedResponse[[]*Menu], error) {
	excludeClause := ` AND NOT (m.branch_id IS NULL AND COALESCE(o.is_excluded, FALSE) = TRUE)`
	return r.listBranchMenus(ctx, filter, branchID, excludeClause)
}

func (r *menuRepository) listBranchManage(ctx context.Context, filter common.Filter, branchID string) (*common.PaginatedResponse[[]*Menu], error) {
	return r.listBranchMenus(ctx, filter, branchID, "")
}

func (r *menuRepository) listBranchMenus(ctx context.Context, filter common.Filter, branchID, extraWhere string) (*common.PaginatedResponse[[]*Menu], error) {
	searchClause := ""
	args := []any{branchID}
	if filter.Search != "" {
		searchClause = " AND m.name ILIKE $2 ESCAPE '\\'"
		args = append(args, "%"+filter.Search+"%")
	}

	page, limit := filter.PageLimit()
	offset := (page - 1) * limit
	argN := len(args)

	query := fmt.Sprintf(`
		SELECT m.id, m.name, m.image, m.deleted_at, m.is_deleted, m.branch_id, m.merchant_id,
			m.is_fasting,
			COALESCE(o.is_available, m.is_available) AS is_available,
			m.description, m.price, m.ingredients, m.category_id, m.modifiers, m.preparation_time,
			m.created_at, m.updated_at,
			COALESCE(o.is_excluded, FALSE) AS is_excluded
		FROM menus m
		INNER JOIN branches b ON b.id = $1 AND b.is_deleted = FALSE
		LEFT JOIN branch_menu_overrides o
			ON o.menu_id = m.id AND o.branch_id = $1 AND o.is_deleted = FALSE
		WHERE m.is_deleted = FALSE
			AND (
				m.branch_id = $1
				OR (m.branch_id IS NULL AND m.merchant_id = b.merchant_id)
			)
			%s%s
		ORDER BY (m.branch_id IS NULL) DESC, m.created_at DESC
		LIMIT $%d OFFSET $%d`, searchClause, extraWhere, argN+1, argN+2)

	args = append(args, limit, offset)

	rows, err := r.listBranchMenuRows(ctx, query, args, false)
	if err != nil {
		r.logger.Error("failed to list branch menus", "error", err)
		return nil, common.ErrInternalServerError
	}

	ptrs := make([]*Menu, len(rows))
	for i := range rows {
		rows[i].Menu.Excluded = rows[i].IsExcluded
		if rows[i].Menu.IsMaster() {
			rows[i].Menu.MasterItem = true
		}
		ptrs[i] = &rows[i].Menu
	}
	return &common.PaginatedResponse[[]*Menu]{
		Data: ptrs,
		Meta: common.BuildPaginationMeta(int64(len(ptrs)), page, limit),
	}, nil
}

type menuWithMeta struct {
	Menu       Menu
	IsExcluded bool
	IsMaster   bool
}

func (r *menuRepository) listBranchMenuRows(ctx context.Context, query string, args []any, withMasterFlag bool) ([]menuWithMeta, error) {
	rawRows, err := common.QueryRows(r.join, ctx, query, args, func(rows *sql.Rows) (menuWithMeta, error) {
		var item menuWithMeta
		scanTargets := []any{
			&item.Menu.ID, &item.Menu.Name, &item.Menu.Image, &item.Menu.DeletedAt, &item.Menu.IsDeleted,
			&item.Menu.BranchID, &item.Menu.MerchantID, &item.Menu.IsFasting, &item.Menu.IsAvailable,
			&item.Menu.Description, &item.Menu.Price, &item.Menu.Ingredients, &item.Menu.CategoryID,
			&item.Menu.Modifiers, &item.Menu.PreparationTime, &item.Menu.CreatedAt, &item.Menu.UpdatedAt,
		}
		if withMasterFlag {
			scanTargets = append(scanTargets, &item.IsMaster)
		}
		scanTargets = append(scanTargets, &item.IsExcluded)
		if err := rows.Scan(scanTargets...); err != nil {
			return menuWithMeta{}, err
		}
		return item, nil
	})
	if err != nil {
		return nil, err
	}
	return rawRows, nil
}

func (r *menuRepository) listAllBranchesEffective(ctx context.Context, filter common.Filter, merchantID string) (*common.PaginatedResponse[[]*Menu], error) {
	searchClause := ""
	args := []any{}
	argIdx := 1
	if merchantID != "" {
		searchClause += fmt.Sprintf(" AND b.merchant_id = $%d", argIdx)
		args = append(args, merchantID)
		argIdx++
	}
	if filter.Search != "" {
		searchClause += fmt.Sprintf(" AND m.name ILIKE $%d ESCAPE '\\'", argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	page, limit := filter.PageLimit()
	offset := (page - 1) * limit

	query := fmt.Sprintf(`
		SELECT m.id, m.name, m.image, m.deleted_at, m.is_deleted,
			COALESCE(m.branch_id, b.id) AS branch_id,
			m.merchant_id,
			m.is_fasting,
			COALESCE(o.is_available, m.is_available) AS is_available,
			m.description, m.price, m.ingredients, m.category_id, m.modifiers, m.preparation_time,
			m.created_at, m.updated_at,
			(m.branch_id IS NULL) AS is_master,
			COALESCE(o.is_excluded, FALSE) AS is_excluded
		FROM branches b
		INNER JOIN menus m ON m.is_deleted = FALSE
			AND (m.branch_id = b.id OR (m.branch_id IS NULL AND m.merchant_id = b.merchant_id))
		LEFT JOIN branch_menu_overrides o
			ON o.menu_id = m.id AND o.branch_id = b.id AND o.is_deleted = FALSE
		WHERE b.is_deleted = FALSE
			AND NOT (m.branch_id IS NULL AND COALESCE(o.is_excluded, FALSE) = TRUE)
			%s
		ORDER BY b.branch_name, m.created_at DESC
		LIMIT $%d OFFSET $%d`, searchClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.listBranchMenuRows(ctx, query, args, true)
	if err != nil {
		r.logger.Error("failed to list all branches effective menus", "error", err)
		return nil, common.ErrInternalServerError
	}

	ptrs := make([]*Menu, len(rows))
	for i := range rows {
		rows[i].Menu.Excluded = rows[i].IsExcluded
		rows[i].Menu.MasterItem = rows[i].IsMaster
		ptrs[i] = &rows[i].Menu
	}
	return &common.PaginatedResponse[[]*Menu]{
		Data: ptrs,
		Meta: common.BuildPaginationMeta(int64(len(ptrs)), page, limit),
	}, nil
}

func scanMenuRow(rows *sql.Rows) (Menu, error) {
	var menu Menu
	if err := rows.Scan(menu.Addr()...); err != nil {
		return Menu{}, err
	}
	return menu, nil
}

func (r *menuRepository) CheckExists(ctx context.Context, name, branchID string) error {
	filter := map[string]any{"name": name, "branch_id": branchID}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check if menu exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrMenuAlreadyExists
}

func (r *menuRepository) CheckMasterExists(ctx context.Context, name, merchantID string) error {
	filter := map[string]any{
		"name":        name,
		"branch_id":   common.IsNull{},
		"merchant_id": merchantID,
		"is_deleted":  false,
	}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check if master menu exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrMenuAlreadyExists
}

func (r *menuRepository) SetBranchOverride(ctx context.Context, branchID, menuID string, isAvailable bool) error {
	query := `
		INSERT INTO branch_menu_overrides (id, branch_id, menu_id, is_available, is_excluded, is_deleted)
		VALUES (uuid_generate_v4(), $1, $2, $3, FALSE, FALSE)
		ON CONFLICT (branch_id, menu_id)
		DO UPDATE SET is_available = EXCLUDED.is_available, is_deleted = FALSE, updated_at = NOW()`

	_, err := r.join.Exec(ctx, query, branchID, menuID, isAvailable)
	if err != nil {
		r.logger.Error("failed to set branch menu override", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *menuRepository) SetBranchExcluded(ctx context.Context, branchID, menuID string, excluded bool) error {
	query := `
		INSERT INTO branch_menu_overrides (id, branch_id, menu_id, is_available, is_excluded, is_deleted)
		VALUES (uuid_generate_v4(), $1, $2, NOT $3, $3, FALSE)
		ON CONFLICT (branch_id, menu_id)
		DO UPDATE SET
			is_excluded = EXCLUDED.is_excluded,
			is_available = CASE
				WHEN EXCLUDED.is_excluded THEN FALSE
				ELSE COALESCE(
					branch_menu_overrides.is_available,
					(SELECT m.is_available FROM menus m WHERE m.id = branch_menu_overrides.menu_id AND m.is_deleted = FALSE),
					TRUE
				)
			END,
			is_deleted = FALSE,
			updated_at = NOW()`

	_, err := r.join.Exec(ctx, query, branchID, menuID, excluded)
	if err != nil {
		r.logger.Error("failed to set branch menu excluded", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *menuRepository) AssignOrphanMasterMenus(ctx context.Context, merchantID string) error {
	if merchantID == "" {
		return nil
	}
	query := `
		UPDATE menus
		SET merchant_id = $1, updated_at = NOW()
		WHERE branch_id IS NULL AND merchant_id IS NULL AND is_deleted = FALSE`
	_, err := r.join.Exec(ctx, query, merchantID)
	if err != nil {
		r.logger.Error("failed to assign orphan master menus to merchant", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *menuRepository) RemoveBranchOverride(ctx context.Context, branchID, menuID string) error {
	query := `
		UPDATE branch_menu_overrides
		SET is_deleted = TRUE, updated_at = NOW()
		WHERE branch_id = $1 AND menu_id = $2 AND is_deleted = FALSE`

	_, err := r.join.Exec(ctx, query, branchID, menuID)
	if err != nil {
		r.logger.Error("failed to remove branch menu override", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}


// menuPublicContextByReference resolves simplified table, branch, and merchant from a QR reference.
const menuPublicContextByReference = `
SELECT
	json_strip_nulls(json_build_object(
		'table_name', t.table_name,
		'reference', t.reference,
		'status', t.status,
		'active_order_id', NULLIF(t.active_order_id::text, '')
	)) AS table,
	json_strip_nulls(json_build_object(
		'id', b.id::text,
		'branch_name', b.branch_name,
		'address', NULLIF(b.address, ''),
		'phone_number', NULLIF(b.phone_number, '')
	)) AS branch,
	json_strip_nulls(json_build_object('name', mer.name, 'logo', NULLIF(mer.logo, ''))) AS merchant
FROM tables t
INNER JOIN branches b ON b.id = t.branch_id AND b.is_deleted = FALSE
INNER JOIN merchants mer ON mer.id = b.merchant_id AND mer.is_deleted = FALSE
WHERE t.reference = $1 AND t.is_deleted = FALSE`

// menuPublicListBase lists available branch-effective menus; CTE resolves branch once, JOINs replace per-row subqueries.
const menuPublicListBase = `
WITH ctx AS (
	SELECT t.branch_id, b.merchant_id
	FROM tables t
	INNER JOIN branches b ON b.id = t.branch_id AND b.is_deleted = FALSE
	WHERE t.reference = $1 AND t.is_deleted = FALSE
)
SELECT
	m.id,
	m.name,
	m.image,
	m.is_fasting,
	COALESCE(o.is_available, m.is_available) AS is_available,
	m.description,
	m.price,
	m.preparation_time,
	CASE WHEN c.id IS NOT NULL THEN json_strip_nulls(json_build_object('name', c.name, 'icon', NULLIF(c.icon, ''))) END AS category,
	COALESCE(ing.ingredients, '[]'::json) AS ingredients,
	COALESCE(mods.modifier_groups, '[]'::json) AS modifier_groups
FROM ctx
INNER JOIN menus m ON m.is_deleted = FALSE
	AND (m.branch_id = ctx.branch_id OR (m.branch_id IS NULL AND m.merchant_id = ctx.merchant_id))
LEFT JOIN branch_menu_overrides o
	ON o.menu_id = m.id AND o.branch_id = ctx.branch_id AND o.is_deleted = FALSE
LEFT JOIN categories c ON c.id = m.category_id AND c.is_deleted = FALSE
LEFT JOIN LATERAL (
	SELECT json_agg(json_strip_nulls(json_build_object('name', i.name, 'icon', NULLIF(i.icon, ''))) ORDER BY u.ord) AS ingredients
	FROM unnest(m.ingredients) WITH ORDINALITY AS u(ing_id, ord)
	INNER JOIN ingredients i ON i.id::text = u.ing_id AND i.is_deleted = FALSE
) ing ON TRUE
LEFT JOIN LATERAL (
	SELECT json_agg(
		json_strip_nulls(json_build_object(
			'id', mg.id,
			'name', mg.name,
			'selection_type', mg.selection_type,
			'is_required', mg.is_required,
			'min_selections', NULLIF(mg.min_selections, 0),
			'max_selections', NULLIF(mg.max_selections, 0),
			'options', COALESCE(opts.options, '[]'::json)
		)) ORDER BY g.ord
	) AS modifier_groups
	FROM unnest(m.modifiers) WITH ORDINALITY AS g(mg_id, ord)
	INNER JOIN modifier_groups mg ON mg.id::text = g.mg_id AND mg.is_deleted = FALSE
	LEFT JOIN LATERAL (
		SELECT json_agg(
			json_strip_nulls(json_build_object(
				'id', mo.id,
				'name', mo.name,
				'price_adjustment', mo.price_adjustment,
				'is_default', CASE WHEN mo.is_default THEN TRUE END,
				'is_available', mo.is_available
			)) ORDER BY o.ord
		) AS options
		FROM unnest(mg.options) WITH ORDINALITY AS o(opt_id, ord)
		INNER JOIN modifier_options mo ON mo.id::text = o.opt_id AND mo.is_deleted = FALSE
	) opts ON TRUE
) mods ON TRUE
WHERE NOT (m.branch_id IS NULL AND COALESCE(o.is_excluded, FALSE) = TRUE)
	AND COALESCE(o.is_available, m.is_available) = TRUE`

// menuPublicCategoriesByReference lists distinct categories used by available menus at the table's branch.
const menuPublicCategoriesByReference = `
WITH ctx AS (
	SELECT t.branch_id, b.merchant_id
	FROM tables t
	INNER JOIN branches b ON b.id = t.branch_id AND b.is_deleted = FALSE
	WHERE t.reference = $1 AND t.is_deleted = FALSE
)
SELECT DISTINCT c.name, c.icon
FROM ctx
INNER JOIN menus m ON m.is_deleted = FALSE
	AND (m.branch_id = ctx.branch_id OR (m.branch_id IS NULL AND m.merchant_id = ctx.merchant_id))
LEFT JOIN branch_menu_overrides o
	ON o.menu_id = m.id AND o.branch_id = ctx.branch_id AND o.is_deleted = FALSE
INNER JOIN categories c ON c.id = m.category_id AND c.is_deleted = FALSE
WHERE NOT (m.branch_id IS NULL AND COALESCE(o.is_excluded, FALSE) = TRUE)
	AND COALESCE(o.is_available, m.is_available) = TRUE
ORDER BY c.name`

func (r *menuRepository) listPublicCategories(ctx context.Context, reference string) ([]*category.CategoryResponseSimplified, error) {
	rows, err := common.QueryRows(r.join, ctx, menuPublicCategoriesByReference, []any{reference}, func(rows *sql.Rows) (*category.CategoryResponseSimplified, error) {
		var cat category.CategoryResponseSimplified
		var icon sql.NullString
		if err := rows.Scan(&cat.Name, &icon); err != nil {
			return nil, err
		}
		if icon.Valid && icon.String != "" {
			cat.Icon = icon.String
		}
		return &cat, nil
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []*category.CategoryResponseSimplified{}, nil
	}
	return rows, nil
}

func (r *menuRepository) getPublicMenuContext(ctx context.Context, reference string) (*PublicMenuCatalogResponse, error) {
	var tableJSON, branchJSON, merchantJSON []byte
	err := r.join.QueryRow(ctx, menuPublicContextByReference, []any{reference}, func(row *sql.Row) error {
		return row.Scan(&tableJSON, &branchJSON, &merchantJSON)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrReferenceNotValid
		}
		return nil, err
	}

	var catalog PublicMenuCatalogResponse
	if err := json.Unmarshal(tableJSON, &catalog.Table); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(branchJSON, &catalog.Branch); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(merchantJSON, &catalog.Merchant); err != nil {
		return nil, err
	}
	return &catalog, nil
}

func (r *menuRepository) scanMenuPublicFromRows(rows *sql.Rows, dto *MenuDTOPublic) error {
	var categoryJSON, ingredientsJSON, modifiersJSON []byte
	var image, description sql.NullString
	err := rows.Scan(
		&dto.ID, &dto.Name, &image,
		&dto.IsFasting, &dto.IsAvailable,
		&description, &dto.Price, &dto.PreparationTime,
		&categoryJSON, &ingredientsJSON, &modifiersJSON,
	)
	if err != nil {
		return err
	}
	if image.Valid && image.String != "" {
		dto.Image = image.String
	}
	if description.Valid && description.String != "" {
		dto.Description = description.String
	}
	return r.unmarshalMenuPublicRelations(dto, categoryJSON, ingredientsJSON, modifiersJSON)
}

func (r *menuRepository) unmarshalMenuPublicRelations(
	dto *MenuDTOPublic,
	categoryJSON, ingredientsJSON, modifiersJSON []byte,
) error {
	if len(categoryJSON) > 0 && string(categoryJSON) != "null" {
		var cat category.CategoryResponseSimplified
		if err := json.Unmarshal(categoryJSON, &cat); err != nil {
			return err
		}
		dto.Category = &cat
	}

	ingredients := []ingredient.IngredientResponseSimplified{}
	if len(ingredientsJSON) > 0 && string(ingredientsJSON) != "null" {
		if err := json.Unmarshal(ingredientsJSON, &ingredients); err != nil {
			return err
		}
	}
	dto.Ingredients = ingredients

	modifiers := []ModifierGroupPublic{}
	if len(modifiersJSON) > 0 && string(modifiersJSON) != "null" {
		if err := json.Unmarshal(modifiersJSON, &modifiers); err != nil {
			return err
		}
	}
	dto.Modifiers = modifiers

	return nil
}

func (r *menuRepository) ListMenus(ctx context.Context, filter common.Filter, reference string) (*PublicMenuCatalogResponse, error) {
	catalog, err := r.getPublicMenuContext(ctx, reference)
	if err != nil {
		if errors.Is(err, common.ErrReferenceNotValid) {
			return nil, err
		}
		r.logger.Error("failed to resolve table reference", "error", err)
		return nil, common.ErrInternalServerError
	}

	categories, err := r.listPublicCategories(ctx, reference)
	if err != nil {
		r.logger.Error("failed to list public categories", "error", err)
		return nil, common.ErrInternalServerError
	}
	catalog.Categories = categories

	common.NormalizeFilter(&filter)
	page, limit := filter.PageLimit()
	offset := (page - 1) * limit

	searchClause := ""
	args := []any{reference}
	if filter.Search != "" {
		searchClause = " AND m.name ILIKE $2 ESCAPE '\\'"
		args = append(args, "%"+filter.Search+"%")
	}
	argN := len(args)
	query := fmt.Sprintf(
		menuPublicListBase+searchClause+` ORDER BY (m.branch_id IS NULL) DESC, m.created_at DESC LIMIT $%d OFFSET $%d`,
		argN+1, argN+2,
	)
	args = append(args, limit, offset)

	menus, err := common.QueryRows(r.join, ctx, query, args, func(rows *sql.Rows) (*MenuDTOPublic, error) {
		var dto MenuDTOPublic
		if err := r.scanMenuPublicFromRows(rows, &dto); err != nil {
			return nil, err
		}
		return &dto, nil
	})
	if err != nil {
		r.logger.Error("failed to list public menus", "error", err)
		return nil, common.ErrInternalServerError
	}

	catalog.Menus = menus
	catalog.Meta = common.BuildPaginationMeta(int64(len(menus)), page, limit)
	return catalog, nil
}
