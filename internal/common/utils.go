package common

import (
	"database/sql"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

func ToNUllString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func ToNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}

func ToNullTimePtr(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func ParseID(r *http.Request, parm string) string {
	return chi.URLParam(r, parm)
}

func normalizePhoneNumber(phoneNumber string) string {
	nonNumericRegex := regexp.MustCompile(`[^0-9]`)
	// 1. Strip everything except digits in one pass
	clean := nonNumericRegex.ReplaceAllString(phoneNumber, "")

	// 2. Handle empty strings to prevent index panics
	if clean == "" {
		return ""
	}

	// 3. Normalize local formats to 251
	if strings.HasPrefix(clean, "0") {
		return "251" + clean[1:]
	}

	// 4. If they start with 9 or 7, assume they forgot the prefix
	if len(clean) == 9 && (clean[0] == '9' || clean[0] == '7') {
		return "251" + clean
	}

	return clean
}

func ValidatePhoneNumber(phoneNumber string) (string, error) {
	normalizedPhoneNumber := normalizePhoneNumber(phoneNumber)
	if normalizedPhoneNumber == "" {
		return "", ErrPhoneNumberRequired
	}

	if len(normalizedPhoneNumber) != 12 {
		return "", ErrInvalidPhoneNumber
	}

	re := regexp.MustCompile(`^251[79]\d{8}$`)
	if !re.MatchString(normalizedPhoneNumber) {
		return "", ErrInvalidPhoneNumber
	}
	return normalizedPhoneNumber, nil
}
