package clientsession

import (
	"lazeez-core/internal/common"
	"time"
)

type ClientSession struct {
	common.Base
	SessionKey     string    `json:"session_key" db:"session_key"`
	TableReference string    `json:"table_reference" db:"table_reference"`
	TableID        string    `json:"table_id" db:"table_id"`
	BranchID       string    `json:"branch_id" db:"branch_id"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
}

func (s *ClientSession) Table() string {
	return "client_sessions"
}

func (s *ClientSession) Columns() []string {
	return []string{
		"id",
		"session_key",
		"table_reference",
		"table_id",
		"branch_id",
		"expires_at",
		"deleted_at",
		"is_deleted",
	}
}

func (s *ClientSession) Values() []any {
	return []any{
		s.ID,
		s.SessionKey,
		s.TableReference,
		s.TableID,
		s.BranchID,
		s.ExpiresAt,
		s.DeletedAt,
		s.IsDeleted,
	}
}

func (s *ClientSession) Addr() []any {
	return []any{
		&s.ID,
		&s.SessionKey,
		&s.TableReference,
		&s.TableID,
		&s.BranchID,
		&s.ExpiresAt,
		&s.DeletedAt,
		&s.IsDeleted,
		&s.CreatedAt,
		&s.UpdatedAt,
	}
}

func (s *ClientSession) IsExpired(now time.Time) bool {
	return !s.ExpiresAt.After(now)
}
