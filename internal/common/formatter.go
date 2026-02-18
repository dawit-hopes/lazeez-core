package common

import (
	"database/sql"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

func FormatText(text string) string {
	if text == "" {
		return text
	}

	r := []rune(strings.ToLower(text))
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}

func ParseStringToUUID(text string) uuid.UUID {
	return uuid.MustParse(text)
}

func ParseUUIDToString(uuid uuid.UUID) string {
	return uuid.String()
}
