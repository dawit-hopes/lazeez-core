package session

import "lazeez-core/internal/common"

type Session struct {
	common.Base
	UserID       string `json:"user_id" db:"user_id"`
	RefreshToken string `json:"refresh_token" db:"refresh_token"`
	AccessToken  string `json:"access_token" db:"access_token"`
	IsRevoked    bool   `json:"is_revoked" db:"is_revoked"`
}

func (s *Session) Table() string {
	return "sessions"
}

func (s *Session) Columns() []string {
	// Note: created_at and updated_at are managed via common.Base and DAL,
	// so we only list table columns that are not part of Base's dynamic handling.
	return []string{"id", "user_id", "refresh_token", "access_token", "is_revoked", "deleted_at", "is_deleted"}
}

func (s *Session) Values() []any {
	return []any{s.ID, s.UserID, s.RefreshToken, s.AccessToken, s.IsRevoked, s.DeletedAt, s.IsDeleted}
}

func (s *Session) Addr() []any {
	return []any{&s.ID, &s.UserID, &s.RefreshToken, &s.AccessToken, &s.IsRevoked, &s.DeletedAt, &s.IsDeleted, &s.CreatedAt, &s.UpdatedAt}
}
