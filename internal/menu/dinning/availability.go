package menu

import (
	"context"
	"database/sql"

	"lazeez-core/internal/common"

	"github.com/lib/pq"
)

// BranchMenuSnapshot is the branch-effective menu state used when validating orders.
type BranchMenuSnapshot struct {
	ID          string
	Name        string
	Price       float64
	IsAvailable bool
	IsExcluded  bool
	// Station is the effective preparation station (menu override or inherited from category).
	Station string
}

const getBranchMenuSnapshotsQuery = `
SELECT m.id::text, m.name, m.price,
	COALESCE(o.is_available, m.is_available) AS is_available,
	COALESCE(o.is_excluded, FALSE) AS is_excluded,
	COALESCE(NULLIF(m.station, ''), c.station, '') AS station
FROM menus m
INNER JOIN branches b ON b.id = $1 AND b.is_deleted = FALSE
LEFT JOIN branch_menu_overrides o
	ON o.menu_id = m.id AND o.branch_id = $1 AND o.is_deleted = FALSE
LEFT JOIN categories c ON c.id = m.category_id AND c.is_deleted = FALSE
WHERE m.is_deleted = FALSE
	AND m.id = ANY($2)
	AND (
		m.branch_id = $1
		OR (m.branch_id IS NULL AND m.merchant_id = b.merchant_id)
	)`

func (r *menuRepository) GetBranchMenuSnapshots(ctx context.Context, branchID string, menuIDs []string) (map[string]BranchMenuSnapshot, error) {
	if len(menuIDs) == 0 {
		r.logger.Debug("branch menu snapshot lookup skipped", "branch_id", branchID, "reason", "empty menu id list")
		return map[string]BranchMenuSnapshot{}, nil
	}

	r.logger.Debug(
		"loading branch menu snapshots",
		"branch_id", branchID,
		"menu_item_count", len(menuIDs),
	)

	rows, err := common.QueryRows(r.join, ctx, getBranchMenuSnapshotsQuery, []any{branchID, pq.Array(menuIDs)}, func(rows *sql.Rows) (BranchMenuSnapshot, error) {
		var snap BranchMenuSnapshot
		if err := rows.Scan(&snap.ID, &snap.Name, &snap.Price, &snap.IsAvailable, &snap.IsExcluded, &snap.Station); err != nil {
			return BranchMenuSnapshot{}, err
		}
		return snap, nil
	})
	if err != nil {
		r.logger.Error(
			"failed to get branch menu snapshots",
			"branch_id", branchID,
			"menu_item_count", len(menuIDs),
			"error", err,
		)
		return nil, common.ErrInternalServerError
	}

	result := make(map[string]BranchMenuSnapshot, len(rows))
	for _, snap := range rows {
		result[snap.ID] = snap
	}

	if len(result) < len(menuIDs) {
		r.logger.Debug(
			"branch menu snapshot lookup returned partial results",
			"branch_id", branchID,
			"requested_count", len(menuIDs),
			"found_count", len(result),
		)
	}

	r.logger.Debug(
		"loaded branch menu snapshots",
		"branch_id", branchID,
		"requested_count", len(menuIDs),
		"found_count", len(result),
	)
	return result, nil
}

func (s *menuService) GetBranchMenuSnapshots(ctx context.Context, branchID string, menuIDs []string) (map[string]BranchMenuSnapshot, error) {
	return s.menuRepository.GetBranchMenuSnapshots(ctx, branchID, menuIDs)
}
