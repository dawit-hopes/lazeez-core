package common

import "database/sql"

func ToNUllString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
