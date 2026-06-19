package feedback

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"lazeez-core/internal/common"
)

type RatingFilter struct {
	common.Filter
	DateFrom   *time.Time
	DateTo     *time.Time
	MinRating  int
	MaxRating  int
	BranchID   string
	MerchantID string
}

func ParseRatingFilter(r *http.Request) RatingFilter {
	f := RatingFilter{Filter: common.ParseFilter(r)}
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
	if v := strings.TrimSpace(query.Get("min_rating")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.MinRating = n
		}
	}
	if v := strings.TrimSpace(query.Get("max_rating")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.MaxRating = n
		}
	}
	if v := strings.TrimSpace(query.Get("branch_id")); v != "" {
		f.BranchID = v
	}
	return f
}

func ValidateRatingDateRange(filter RatingFilter) error {
	if filter.DateFrom != nil && filter.DateTo != nil {
		fromDay := filter.DateFrom.Truncate(24 * time.Hour)
		toDay := filter.DateTo.Truncate(24 * time.Hour)
		if fromDay.After(toDay) {
			return common.ErrInvalidRequest
		}
	}
	return nil
}

func buildOrderRatingFilterClause(filter RatingFilter, baseArgs []any) (string, []any) {
	return buildRatingFilterClause(filter, baseArgs, "r", true)
}

func buildStayRatingFilterClause(filter RatingFilter, baseArgs []any) (string, []any) {
	return buildRatingFilterClause(filter, baseArgs, "r", false)
}

func buildRatingFilterClause(filter RatingFilter, baseArgs []any, alias string, includeOrderNumber bool) (string, []any) {
	args := append([]any{}, baseArgs...)
	conds := make([]string, 0, 8)
	n := len(args) + 1

	if filter.DateFrom != nil {
		conds = append(conds, alias+".created_at >= $"+strconv.Itoa(n))
		args = append(args, *filter.DateFrom)
		n++
	}
	if filter.DateTo != nil {
		conds = append(conds, alias+".created_at <= $"+strconv.Itoa(n))
		args = append(args, *filter.DateTo)
		n++
	}
	if filter.MinRating > 0 {
		conds = append(conds, alias+".rating >= $"+strconv.Itoa(n))
		args = append(args, filter.MinRating)
		n++
	}
	if filter.MaxRating > 0 {
		conds = append(conds, alias+".rating <= $"+strconv.Itoa(n))
		args = append(args, filter.MaxRating)
		n++
	}
	if filter.BranchID != "" {
		conds = append(conds, alias+".branch_id = $"+strconv.Itoa(n))
		args = append(args, filter.BranchID)
		n++
	}
	if filter.MerchantID != "" {
		conds = append(conds, alias+".branch_id IN (SELECT b.id FROM branches b WHERE b.merchant_id = $"+strconv.Itoa(n)+" AND b.is_deleted = FALSE)")
		args = append(args, filter.MerchantID)
		n++
	}
	if filter.Search != "" {
		pattern := common.ILikePattern(filter.Search)
		if includeOrderNumber {
			conds = append(conds, fmt.Sprintf(
				"(%s.comment ILIKE $%d OR COALESCE(%s.phone_number, '') ILIKE $%d OR %s.session_key ILIKE $%d OR %s.order_id::text ILIKE $%d OR COALESCE(o.order_number::text, '') ILIKE $%d OR EXISTS (SELECT 1 FROM unnest(%s.tags) t(tag) WHERE t.tag ILIKE $%d))",
				alias, n, alias, n+1, alias, n+2, alias, n+3, n+4, alias, n+5,
			))
		} else {
			conds = append(conds, fmt.Sprintf(
				"(%s.comment ILIKE $%d OR COALESCE(%s.phone_number, '') ILIKE $%d OR %s.session_key ILIKE $%d OR COALESCE(%s.table_name, '') ILIKE $%d OR EXISTS (SELECT 1 FROM unnest(%s.tags) t(tag) WHERE t.tag ILIKE $%d))",
				alias, n, alias, n+1, alias, n+2, alias, n+3, alias, n+4,
			))
		}
		placeholders := 6
		if !includeOrderNumber {
			placeholders = 5
		}
		for i := 0; i < placeholders; i++ {
			args = append(args, pattern)
		}
	}

	if len(conds) == 0 {
		return "", args
	}
	return " AND " + strings.Join(conds, " AND "), args
}
