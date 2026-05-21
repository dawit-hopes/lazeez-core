package clientsession

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"time"
)

const sessionDuration = time.Hour

type ClientSessionService interface {
	Create(ctx context.Context, input CreateSessionInput) (*ClientSessionResponse, error)
	GetValidForOrder(ctx context.Context, sessionKey string) (*ClientSession, error)
}

type clientSessionService struct {
	repo   ClientSessionRepository
	logger config.Logger
}

func NewClientSessionService(repo ClientSessionRepository, logger config.Logger) ClientSessionService {
	return &clientSessionService{repo: repo, logger: logger}
}

func (s *clientSessionService) Create(ctx context.Context, input CreateSessionInput) (*ClientSessionResponse, error) {
	table, err := s.repo.GetTableByReference(ctx, input.Reference)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &ClientSession{
		SessionKey:     common.GenerateUUID(),
		TableReference: input.Reference,
		TableID:        table.ID,
		BranchID:       table.BranchID,
		ExpiresAt:      now.Add(sessionDuration),
	}
	session.ID = common.GenerateUUID()

	if err := s.repo.Create(ctx, session); err != nil {
		s.logger.Error("failed to create client session", "error", err)
		return nil, common.ErrInternalServerError
	}

	return session.ToResponse(table.TableName), nil
}

func (s *clientSessionService) GetValidForOrder(ctx context.Context, sessionKey string) (*ClientSession, error) {
	session, err := s.repo.GetBySessionKey(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	if session.IsExpired(time.Now()) {
		return nil, common.ErrClientSessionExpired
	}
	return session, nil
}
