package order

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"lazeez-core/internal/common"
)

// OrderFilter extends common.Filter with order-specific filters for branch and admin.
type OrderFilter struct {
	common.Filter
	DateFrom   *time.Time // filter orders from this date (inclusive)
	DateTo     *time.Time // filter orders to this date (inclusive)
	OrderStatus string   // filter by order_status
	BranchID   string    // for super_admin: filter by branch
	MerchantID string    // for super_branch_admin: restrict to merchant branches
}

// ParseOrderFilter parses query params into OrderFilter.
// Supports: page, limit, search, date_from, date_to, order_status, branch_id
func ParseOrderFilter(r *http.Request) OrderFilter {
	f := OrderFilter{Filter: common.ParseFilter(r)}
	applyOrderFilterQuery(&f, r)
	return f
}

// ParseArchiveOrderFilter parses query params for the branch archive endpoint.
// Defaults to page=1, limit=20. Date filters apply to updated_at (last activity).
func ParseArchiveOrderFilter(r *http.Request) OrderFilter {
	f := OrderFilter{Filter: common.ParseFilter(r)}
	if r.URL.Query().Get("limit") == "" {
		f.Limit = common.DefaultPageLimit
	}
	if f.Limit <= 0 {
		f.Limit = common.DefaultPageLimit
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	applyOrderFilterQuery(&f, r)
	return f
}

func applyOrderFilterQuery(f *OrderFilter, r *http.Request) {
	query := r.URL.Query()
	if v := query.Get("date_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.DateFrom = &t
		}
	}
	if v := query.Get("date_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			t = t.Add(24*time.Hour - time.Nanosecond)
			f.DateTo = &t
		}
	}
	if v := strings.TrimSpace(query.Get("order_status")); v != "" {
		f.OrderStatus = v
	}
	if v := strings.TrimSpace(query.Get("branch_id")); v != "" {
		f.BranchID = v
	}
}

// ValidateArchiveDateRange ensures date_from is not after date_to when both are set.
func ValidateArchiveDateRange(filter OrderFilter) error {
	if filter.DateFrom != nil && filter.DateTo != nil {
		fromDay := filter.DateFrom.Truncate(24 * time.Hour)
		toDay := filter.DateTo.Truncate(24 * time.Hour)
		if fromDay.After(toDay) {
			return common.ErrInvalidRequest
		}
	}
	return nil
}

// SessionKeyFromRequest returns session key from X-Session-Key header or session_key query param.
func SessionKeyFromRequest(r *http.Request) string {
	if k := r.Header.Get("X-Session-Key"); k != "" {
		return strings.TrimSpace(k)
	}
	return strings.TrimSpace(r.URL.Query().Get("session_key"))
}

// BuildOrderFilterClause appends filter conditions and returns (clause, args).
// baseArgs are prepended; filter args are appended. Placeholders use $1, $2, ... for the combined slice.
func BuildOrderFilterClause(filter OrderFilter, baseArgs []any) (string, []any) {
	return buildOrderFilterClause(filter, baseArgs, "o.created_at")
}

// BuildArchiveOrderFilterClause filters archive rows by updated_at and optional status/search.
func BuildArchiveOrderFilterClause(filter OrderFilter, baseArgs []any) (string, []any) {
	return buildOrderFilterClause(filter, baseArgs, "o.updated_at")
}

func buildOrderFilterClause(filter OrderFilter, baseArgs []any, dateColumn string) (string, []any) {
	var conds []string
	args := append([]any{}, baseArgs...)
	n := len(baseArgs) + 1

	if filter.DateFrom != nil {
		conds = append(conds, dateColumn+" >= $"+strconv.Itoa(n))
		args = append(args, filter.DateFrom)
		n++
	}
	if filter.DateTo != nil {
		conds = append(conds, dateColumn+" <= $"+strconv.Itoa(n))
		args = append(args, filter.DateTo)
		n++
	}
	if filter.OrderStatus != "" {
		conds = append(conds, "o.order_status = $"+strconv.Itoa(n))
		args = append(args, filter.OrderStatus)
		n++
	}
	if filter.BranchID != "" {
		conds = append(conds, "o.branch_id = $"+strconv.Itoa(n))
		args = append(args, filter.BranchID)
		n++
	}
	if filter.MerchantID != "" {
		conds = append(conds, "o.branch_id IN (SELECT b.id FROM branches b WHERE b.merchant_id = $"+strconv.Itoa(n)+" AND b.is_deleted = FALSE)")
		args = append(args, filter.MerchantID)
		n++
	}
	if filter.Search != "" {
		pattern := common.ILikePattern(filter.Search)
		conds = append(conds, "(o.order_number::text ILIKE $"+strconv.Itoa(n)+" OR o.table_number::text ILIKE $"+strconv.Itoa(n)+" OR o.id::text ILIKE $"+strconv.Itoa(n)+" OR o.session_key ILIKE $"+strconv.Itoa(n)+")")
		args = append(args, pattern)
	}

	if len(conds) == 0 {
		return "", args
	}
	return " AND " + strings.Join(conds, " AND "), args
}
