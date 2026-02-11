package auth

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/key"
	"lazeez-core/internal/session"
	"lazeez-core/internal/users"
)

var (
	accessTokenExpirationMinutes  = 15
	refreshTokenExpirationMinutes = 60 * 24 * 30
)

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
	FirstTimeLogin(ctx context.Context, req SetPasswordRequest) (*LoginResponse, error)
	ResetPassword(ctx context.Context, req SetPasswordRequest) (*LoginResponse, error)
	Logout(ctx context.Context, id string) error
}

type authService struct {
	userService    users.UserService
	sessionService session.SessionService
	keyService     key.KeyService
	logger         config.Logger
}

func NewAuthService(userService users.UserService, sessionService session.SessionService, keyService key.KeyService, logger config.Logger) AuthService {
	return &authService{
		userService:    userService,
		sessionService: sessionService,
		keyService:     keyService,
		logger:         logger,
	}
}

func (s *authService) validateUser(ctx context.Context, user *users.User, ok bool) error {
	if !ok {
		s.logger.Error("Invalid password", "phone number", user.PhoneNumber)
		err := s.userService.UpdateLoggingAttempts(ctx, user.ID, user.LoggingAttempts+1)
		if err != nil {
			s.logger.Error("Failed to update logging attempts", "error", err)
			return err
		}
		return common.ErrUnAuthorized
	}

	if user.LoggingAttempts >= 0 {
		err := s.userService.ResetLoggingAttempts(ctx, user.ID)
		if err != nil {
			s.logger.Error("Failed to reset logging attempts", "error", err)
			return err
		}
	}

	if user.IsLocked {
		s.logger.Error("User is locked", "phone number", user.PhoneNumber)
		return common.ErrUserLocked
	}

	if user.LoggingAttempts >= 5 {
		err := s.userService.LockUser(ctx, user.ID)
		if err != nil {
			s.logger.Error("Failed to lock user", "error", err)
			return err
		}
	}
	return nil
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	s.logger.Info("Logging in", "phone number", req.PhoneNumber)
	normalizedPhoneNumber, err := s.userService.ValidatePhoneNumber(req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to validate phone number", "error", err)
		return LoginResponse{}, err
	}

	existingUser, err := s.userService.GetUserByPhoneNumber(ctx, normalizedPhoneNumber)
	if err != nil {
		s.logger.Error("Failed to get user by phone number", "error", err)
		return LoginResponse{}, err
	}
	if existingUser == nil {
		s.logger.Error("User not found", "phone number", req.PhoneNumber)
		return LoginResponse{}, common.ErrUserNotFound
	}
	ok, err := s.keyService.VerifyPassword(req.Password, existingUser.Password)
	if err != nil {
		s.logger.Error("Failed to verify password", "error", err)
		// Treat as invalid password - update attempts and return user-friendly error
		if validateErr := s.validateUser(ctx, existingUser, false); validateErr != nil {
			return LoginResponse{}, validateErr
		}
		return LoginResponse{}, common.ErrUnAuthorized
	}

	if err := s.validateUser(ctx, existingUser, ok); err != nil {
		s.logger.Error("Failed to validate user", "error", err)
		return LoginResponse{}, err
	}

	accessToken, refreshToken, err := s.generateTokens(existingUser)
	if err != nil {
		s.logger.Error("Failed to generate tokens", "error", err)
		return LoginResponse{}, err
	}

	sess := session.Session{
		UserID:       existingUser.ID,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		IsRevoked:    false,
	}

	s.logger.Info("session", "session", sess.UserID)

	sessionErr := s.sessionService.CreateSession(ctx, sess)
	if sessionErr != nil {
		s.logger.Error("Failed to create session", "error", err)
		return LoginResponse{}, err
	}

	return LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         existingUser.ToDTO(),
	}, nil
}

func (s *authService) FirstTimeLogin(ctx context.Context, req SetPasswordRequest) (*LoginResponse, error) {
	return s.setPasswordAndLogin(ctx, req, true)
}

func (s *authService) ResetPassword(ctx context.Context, req SetPasswordRequest) (*LoginResponse, error) {
	return s.setPasswordAndLogin(ctx, req, false)
}

func (s *authService) setPasswordAndLogin(ctx context.Context, req SetPasswordRequest, requireFirstLogin bool) (*LoginResponse, error) {
	normalizedPhoneNumber, err := s.userService.ValidatePhoneNumber(req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to validate phone number", "error", err)
		return nil, err
	}

	existingUser, err := s.userService.GetUserByPhoneNumber(ctx, normalizedPhoneNumber)
	if err != nil {
		s.logger.Error("Failed to get user by phone number", "error", err)
		return nil, err
	}
	if existingUser == nil {
		s.logger.Error("User not found", "phone number", req.PhoneNumber)
		return nil, common.ErrUserNotFound
	}

	if requireFirstLogin && !existingUser.IsFirstLogin {
		s.logger.Error("User is not first time login", "phone number", req.PhoneNumber)
		return nil, common.ErrUserNotFirstTimeLogin
	}
	if !requireFirstLogin && existingUser.IsFirstLogin {
		s.logger.Error("User is first time login", "phone number", req.PhoneNumber)
		return nil, common.ErrUserIsFirstTimeLogin
	}

	err = s.validatePassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to validate password", "error", err)
		return nil, err
	}

	encryptedPassword, err := s.keyService.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, err
	}

	err = s.userService.SetPassword(ctx, req.PhoneNumber, existingUser.ID, encryptedPassword, existingUser.IsFirstLogin)
	if err != nil {
		s.logger.Error("Failed to set password", "error", err)
		return nil, err
	}

	accessToken, refreshToken, err := s.generateTokens(existingUser)
	if err != nil {
		s.logger.Error("Failed to generate tokens", "error", err)
		return nil, err
	}

	return &LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: existingUser.ToDTO()}, nil
}

func (s *authService) Logout(ctx context.Context, id string) error {
	s.logger.Info("Logging out", "id", id)
	err := s.sessionService.RevokeSession(ctx, id, true)
	if err != nil {
		s.logger.Error("Failed to revoke session", "error", err)
		return err
	}
	return nil
}
