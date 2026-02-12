package common

import (
	"database/sql"
	"net/http"
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
