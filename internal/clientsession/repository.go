package clientsession

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type tableContext struct {
	ID        string
	BranchID  string
	TableName string
}

type ClientSessionRepository interface {
	Create(ctx context.Context, session *ClientSession) error
	GetBySessionKey(ctx context.Context, sessionKey string) (*ClientSession, error)
	GetTableByReference(ctx context.Context, reference string) (*tableContext, error)
}

type clientSessionRepository struct {
	dal     *common.DAL[*ClientSession]
	joinDAL *common.JoinDAL
	logger  config.Logger
}

func NewClientSessionRepository(dal *common.DAL[*ClientSession], joinDAL *common.JoinDAL, logger config.Logger) ClientSessionRepository {
	return &clientSessionRepository{dal: dal, joinDAL: joinDAL, logger: logger}
}

func (r *clientSessionRepository) Create(ctx context.Context, session *ClientSession) error {
	_, err := r.dal.Create(ctx, session)
	return err
}

func (r *clientSessionRepository) GetBySessionKey(ctx context.Context, sessionKey string) (*ClientSession, error) {
	const query = `
SELECT cs.id, cs.session_key, cs.table_reference, cs.table_id, cs.branch_id, cs.expires_at,
       cs.is_deleted, cs.created_at, cs.updated_at, cs.deleted_at
FROM client_sessions cs
WHERE cs.session_key = $1 AND cs.is_deleted = FALSE
LIMIT 1`

	var session ClientSession
	err := r.joinDAL.QueryRow(ctx, query, []any{sessionKey}, func(row *sql.Row) error {
		return row.Scan(
			&session.ID,
			&session.SessionKey,
			&session.TableReference,
			&session.TableID,
			&session.BranchID,
			&session.ExpiresAt,
			&session.IsDeleted,
			&session.CreatedAt,
			&session.UpdatedAt,
			&session.DeletedAt,
		)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrClientSessionNotFound
		}
		r.logger.Error("failed to get client session", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &session, nil
}

func (r *clientSessionRepository) GetTableByReference(ctx context.Context, reference string) (*tableContext, error) {
	const query = `
SELECT t.id, t.branch_id::text, t.table_name
FROM tables t
WHERE t.reference = $1 AND t.is_deleted = FALSE
LIMIT 1`

	var table tableContext
	err := r.joinDAL.QueryRow(ctx, query, []any{reference}, func(row *sql.Row) error {
		return row.Scan(&table.ID, &table.BranchID, &table.TableName)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrReferenceNotValid
		}
		r.logger.Error("failed to resolve table reference", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &table, nil
}
