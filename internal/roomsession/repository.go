package roomsession

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type RoomSessionRepository interface {
	Create(ctx context.Context, session *RoomSession) error
	GetBySessionKey(ctx context.Context, sessionKey string) (*RoomSession, error)
}

type roomSessionRepository struct {
	dal     *common.DAL[*RoomSession]
	joinDAL *common.JoinDAL
	logger  config.Logger
}

func NewRoomSessionRepository(dal *common.DAL[*RoomSession], joinDAL *common.JoinDAL, logger config.Logger) RoomSessionRepository {
	return &roomSessionRepository{dal: dal, joinDAL: joinDAL, logger: logger}
}

func (r *roomSessionRepository) Create(ctx context.Context, session *RoomSession) error {
	if _, err := r.dal.Create(ctx, session); err != nil {
		r.logger.Error("failed to create room session", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *roomSessionRepository) GetBySessionKey(ctx context.Context, sessionKey string) (*RoomSession, error) {
	filter := map[string]any{"session_key": sessionKey, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrRoomSessionNotFound
		}
		r.logger.Error("failed to get room session", "error", err)
		return nil, common.ErrInternalServerError
	}
	return result, nil
}
