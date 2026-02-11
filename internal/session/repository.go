package session

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type SessionRepository interface {
	Create(ctx context.Context, session Session) (Session, error)
	Get(ctx context.Context, id string) (Session, error)
	GetByUserID(ctx context.Context, userID string) (Session, error)
	Update(ctx context.Context, session Session) (Session, error)
	Delete(ctx context.Context, id string) error
	Revoke(ctx context.Context, id string, action bool) error
}

type sessionRepository struct {
	dal    *common.DAL[*Session]
	logger config.Logger
}

func NewSessionRepository(dal *common.DAL[*Session], logger config.Logger) SessionRepository {
	return &sessionRepository{dal: dal, logger: logger}
}

func (r *sessionRepository) Create(ctx context.Context, session Session) (Session, error) {
	result, err := r.dal.Create(ctx, &session)
	if err != nil {
		r.logger.Error("failed to create session", "error", err)
		return session, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *sessionRepository) Get(ctx context.Context, id string) (Session, error) {
	filter := map[string]any{"user_id": id, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("session not found", "error", err)
			return Session{}, common.ErrSessionNotFound
		}
		r.logger.Error("failed to get session", "error", err)
		return Session{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *sessionRepository) Update(ctx context.Context, session Session) (Session, error) {
	filter := map[string]any{"id": session.ID}
	updates := map[string]any{
		"refresh_token": session.RefreshToken,
		"access_token":  session.AccessToken,
		"is_revoked":    session.IsRevoked,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("session not found", "error", err)
			return Session{}, common.ErrSessionNotFound
		}
		r.logger.Error("failed to update session", "error", err)
		return Session{}, common.ErrInternalServerError
	}

	return session, nil
}

func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	err := r.dal.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("session not found", "error", err)
			return common.ErrSessionNotFound
		}
		r.logger.Error("failed to delete session", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *sessionRepository) GetByUserID(ctx context.Context, userID string) (Session, error) {
	filter := map[string]any{"user_id": userID, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("session not found", "error", err)
			return Session{}, nil
		}
		r.logger.Error("failed to get session", "error", err)
		return Session{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *sessionRepository) Revoke(ctx context.Context, id string, action bool) error {
	updates := map[string]any{"is_revoked": action}
	err := r.dal.Update(ctx, map[string]any{"id": id}, updates)
	if err != nil {
		r.logger.Error("failed to revoke session", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}
