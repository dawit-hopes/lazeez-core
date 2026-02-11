package session

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type SessionService interface {
	CreateSession(ctx context.Context, session Session) error
	GetSession(ctx context.Context, id string) (Session, error)
	UpdateSession(ctx context.Context, session Session) (Session, error)
	DeleteSession(ctx context.Context, id string) error
	RevokeSession(ctx context.Context, id string, action bool) error
}

type sessionService struct {
	sessionRepository SessionRepository
	logger            config.Logger
}

func NewSessionService(sessionRepository SessionRepository, logger config.Logger) SessionService {
	return &sessionService{sessionRepository: sessionRepository, logger: logger}
}

func (s *sessionService) CreateSession(ctx context.Context, session Session) error {
	existingSession, err := s.sessionRepository.GetByUserID(ctx, session.UserID)
	if err != nil {
		s.logger.Error("failed to get session", "error", err)
		return err
	}

	if existingSession.ID != "" {
		existingSession.RefreshToken = session.RefreshToken
		existingSession.AccessToken = session.AccessToken
		existingSession.IsRevoked = false
		_, err = s.sessionRepository.Update(ctx, existingSession)
		if err != nil {
			s.logger.Error("failed to update session", "error", err)
			return err
		}
		s.logger.Info("session updated successfully")
		return nil
	}	

	// No existing session - create new
	session.ID = common.GenerateUUID()
	_, err = s.sessionRepository.Create(ctx, session)
	if err != nil {
		s.logger.Error("failed to create session", "error", err)
		return err
	}

	s.logger.Info("session created successfully")
	return nil
}

func (s *sessionService) GetSession(ctx context.Context, id string) (Session, error) {
	session, err := s.sessionRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("failed to get session", "error", err)
		return Session{}, err
	}
	return session, nil
}

func (s *sessionService) UpdateSession(ctx context.Context, session Session) (Session, error) {
	session, err := s.sessionRepository.Update(ctx, session)
	if err != nil {
		s.logger.Error("failed to update session", "error", err)
		return Session{}, err
	}
	return session, nil
}

func (s *sessionService) DeleteSession(ctx context.Context, id string) error {
	return s.sessionRepository.Delete(ctx, id)
}

func (s *sessionService) RevokeSession(ctx context.Context, id string, action bool) error {
	err := s.sessionRepository.Revoke(ctx, id, action)
	if err != nil {
		s.logger.Error("failed to revoke session", "error", err)
		return err
	}
	return nil
}
