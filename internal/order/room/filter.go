package roomorder

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"lazeez-core/internal/common"
)

// RoomOrderFilter extends common.Filter with room-order-specific filters.
type RoomOrderFilter struct {
	common.Filter
	DateFrom    *time.Time
	DateTo      *time.Time
	OrderStatus string
	BranchID    string
	BookingID   string
	MerchantID  string
}

func ParseRoomOrderFilter(r *http.Request) RoomOrderFilter {
	f := RoomOrderFilter{Filter: common.ParseFilter(r)}
	applyRoomOrderFilterQuery(&f, r)
	return f
}

func ParseArchiveRoomOrderFilter(r *http.Request) RoomOrderFilter {
	f := RoomOrderFilter{Filter: common.ParseFilter(r)}
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
	applyRoomOrderFilterQuery(&f, r)
	return f
}

func applyRoomOrderFilterQuery(f *RoomOrderFilter, r *http.Request) {
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
	if v := strings.TrimSpace(query.Get("booking_id")); v != "" {
		f.BookingID = v
	}
}

func ValidateArchiveDateRange(filter RoomOrderFilter) error {
	if filter.DateFrom != nil && filter.DateTo != nil {
		fromDay := filter.DateFrom.Truncate(24 * time.Hour)
		toDay := filter.DateTo.Truncate(24 * time.Hour)
		if fromDay.After(toDay) {
			return common.ErrInvalidRequest
		}
	}
	return nil
}

// GuestAuthFromRequest returns room reference and passcode from query params or headers.
func GuestAuthFromRequest(r *http.Request) (reference, passCode string, err error) {
	reference = strings.TrimSpace(r.URL.Query().Get("reference"))
	if reference == "" {
		reference = strings.TrimSpace(r.Header.Get("X-Room-Reference"))
	}
	passCode = strings.TrimSpace(r.URL.Query().Get("pass_code"))
	if passCode == "" {
		passCode = strings.TrimSpace(r.Header.Get("X-Pass-Code"))
	}
	if reference == "" || passCode == "" {
		return "", "", common.ErrInvalidRequest
	}
	return reference, passCode, nil
}

func BuildRoomOrderFilterClause(filter RoomOrderFilter, baseArgs []any) (string, []any) {
	return buildRoomOrderFilterClause(filter, baseArgs, "o.created_at")
}

func BuildArchiveRoomOrderFilterClause(filter RoomOrderFilter, baseArgs []any) (string, []any) {
	return buildRoomOrderFilterClause(filter, baseArgs, "o.updated_at")
}

func buildRoomOrderFilterClause(filter RoomOrderFilter, baseArgs []any, dateColumn string) (string, []any) {
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
	if filter.BookingID != "" {
		conds = append(conds, "o.booking_id = $"+strconv.Itoa(n))
		args = append(args, filter.BookingID)
		n++
	}
	if filter.MerchantID != "" {
		conds = append(conds, "o.branch_id IN (SELECT b.id FROM branches b WHERE b.merchant_id = $"+strconv.Itoa(n)+" AND b.is_deleted = FALSE)")
		args = append(args, filter.MerchantID)
		n++
	}
	if filter.Search != "" {
		pattern := common.ILikePattern(filter.Search)
		conds = append(conds, "(o.order_number::text ILIKE $"+strconv.Itoa(n)+" OR o.id::text ILIKE $"+strconv.Itoa(n)+" OR o.session_key ILIKE $"+strconv.Itoa(n)+")")
		args = append(args, pattern)
	}

	if len(conds) == 0 {
		return "", args
	}
	return " AND " + strings.Join(conds, " AND "), args
}
