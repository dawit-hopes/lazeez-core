package auth

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/key"
	"lazeez-core/internal/session"
	"lazeez-core/internal/users"

	"golang.org/x/sync/errgroup"
)

var (
	accessTokenExpirationMinutes  = 15
	refreshTokenExpirationMinutes = 60 * 24 * 30
)

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	FirstTimeLogin(ctx context.Context, req SetPasswordRequest) (*LoginResponse, error)
	ResetPassword(ctx context.Context, req SetPasswordRequest) (*LoginResponse, error)
	Logout(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
}

type authService struct {
	userService    users.UserService
	sessionService session.SessionService
	keyService     key.KeyService
	logger         config.Logger
	JWTSecretKey   string
}

func NewAuthService(userService users.UserService, sessionService session.SessionService, keyService key.KeyService, logger config.Logger, jwtSecretKey string) AuthService {
	return &authService{
		userService:    userService,
		sessionService: sessionService,
		keyService:     keyService,
		logger:         logger,
		JWTSecretKey:   jwtSecretKey,
	}
}

func (s *authService) validateUser(ctx context.Context, user *users.User, req LoginRequest) error {
	isValidPassword, err := s.keyService.VerifyPassword(req.Password, user.Password)
	if err != nil {
		s.logger.Error("Failed to verify password", "error", err)
		updateErr := s.userService.UpdateLoggingAttempts(ctx, user.ID, user.LoggingAttempts+1)
		if updateErr != nil {
			s.logger.Error("Failed to update logging attempts", "error", updateErr)
			return updateErr
		}
		return common.ErrWrongUsernameOrPassword
	}

	if !isValidPassword {
		s.logger.Error("Invalid password", "phone number", user.PhoneNumber)
		updateErr := s.userService.UpdateLoggingAttempts(ctx, user.ID, user.LoggingAttempts+1)
		if updateErr != nil {
			s.logger.Error("Failed to update logging attempts", "error", updateErr)
			return updateErr
		}
		return common.ErrWrongUsernameOrPassword
	}

	if user.LoggingAttempts > 0 {
		err := s.userService.ResetLoggingAttempts(ctx, user.ID)
		if err != nil {
			s.logger.Error("Failed to reset logging attempts", "error", err)
			return err
		}
	}

	// if user.IsLocked {
	// 	s.logger.Error("User is locked", "phone number", user.PhoneNumber)
	// 	return common.ErrUserLocked
	// }

	// if user.LoggingAttempts >= 5 {
	// 	err := s.userService.LockUser(ctx, user.ID)
	// 	if err != nil {
	// 		s.logger.Error("Failed to lock user", "error", err)
	// 		return err
	// 	}
	// }
	return nil
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	s.logger.Info("Logging in", "phone number", req.PhoneNumber)
	// req.PhoneNumber is already validated and normalized by the handler
	existingUser, err := s.userService.GetUserByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to get user by phone number", "error", err)
		return nil, err
	}

	if err := s.validateUser(ctx, existingUser, req); err != nil {
		s.logger.Error("Failed to validate user", "error", err)
		return nil, err
	}

	response, err := s.createLoginResponse(existingUser.ToDTO())
	if err != nil {
		s.logger.Error("Failed to create login response", "error", err)
		return nil, err
	}

	sess := session.Session{
		UserID:       existingUser.ID,
		RefreshToken: response.RefreshToken,
		AccessToken:  response.AccessToken,
		IsRevoked:    false,
	}

	sessionErr := s.sessionService.CreateSession(ctx, sess)
	if sessionErr != nil {
		s.logger.Error("Failed to create session", "error", sessionErr)
		return nil, sessionErr
	}

	return response, nil
}

func (s *authService) FirstTimeLogin(ctx context.Context, req SetPasswordRequest) (*LoginResponse, error) {
	return s.setPasswordAndLogin(ctx, req, true)
}

func (s *authService) ResetPassword(ctx context.Context, req SetPasswordRequest) (*LoginResponse, error) {
	return s.setPasswordAndLogin(ctx, req, false)
}

func (s *authService) setPasswordAndLogin(ctx context.Context, req SetPasswordRequest, requireFirstLogin bool) (*LoginResponse, error) {
	// req.PhoneNumber is already validated and normalized by the handler
	existingUser, err := s.validateExistingUser(ctx, req.PhoneNumber, requireFirstLogin)
	if err != nil {
		s.logger.Error("Failed to validate existing user", "error", err)
		return nil, err
	}
	if err := s.validatePassword(req.Password); err != nil {
		s.logger.Error("Failed to validate password", "error", err)
		return nil, err
	}
	encryptedPassword, err := s.keyService.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, err
	}
	if err := s.userService.SetPassword(ctx, req.PhoneNumber, existingUser.ID, encryptedPassword, existingUser.IsFirstLogin); err != nil {
		s.logger.Error("Failed to set password", "error", err)
		return nil, err
	}

	return s.createLoginResponse(existingUser.ToDTO())
}

func (s *authService) Logout(ctx context.Context, userID string) error {
	s.logger.Info("Logging out", "user ID", userID)
	revokeErr := s.sessionService.RevokeSession(ctx, userID, true)
	if revokeErr != nil {
		s.logger.Error("Failed to revoke session", "error", revokeErr)
		return revokeErr
	}
	return nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	var user users.UserDTO
	data, err := s.keyService.DecodeJWTToken(refreshToken, s.JWTSecretKey)
	if err != nil {
		s.logger.Error("Failed to decode refresh token", "error", err)
		return nil, err
	}

	userID, ok := data["uid"].(string)
	if !ok {
		s.logger.Error("Failed to get user ID from refresh token", "error", err)
		return nil, common.ErrInvalidRefreshToken
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.validateSession(ctx, userID, refreshToken)
	})

	g.Go(func() error {
		result, err := s.userService.GetUserByID(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to get user by ID", "error", err)
			return err
		}
		user = result
		return nil
	})

	if err := g.Wait(); err != nil {
		s.logger.Error("Failed to validate session or get user", "error", err)
		return nil, err
	}

	return s.createLoginResponse(user)
}

func (s *authService) validateSession(ctx context.Context, userID string, refreshToken string) error {
	session, err := s.sessionService.GetSessionByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get session", "error", err)
		return err
	}

	if session.IsRevoked {
		s.logger.Error("Session is revoked", "refresh token", userID)
		return common.ErrSessionRevoked
	}

	if session.RefreshToken != refreshToken {
		s.logger.Error("Refresh token is invalid", "refresh token", userID)
		return common.ErrInvalidRefreshToken
	}
	return nil
}

func (s *authService) generatePayload(user users.UserDTO) map[string]any {
	// This payload is later consumed by generateTokens, which expects:
	//   - "id"        → user ID
	//   - "role"      → user role
	//   - "branch_id" → branch ID (used to populate the "bid" claim in the JWT)
	// Previously we stored "bid" here, which meant generateTokens could not
	// find "branch_id" and was writing an empty "bid" claim into the token.
	return map[string]any{
		"id":   user.ID,
		"role": user.Role,
		"bid":  user.BranchID,
	}
}

func (s *authService) validateExistingUser(ctx context.Context, normalizedPhoneNumber string, requireFirstLogin bool) (*users.User, error) {
	user, err := s.userService.GetUserByPhoneNumber(ctx, normalizedPhoneNumber)
	if err != nil {
		s.logger.Error("Failed to get user by phone number", "error", err)
		return nil, err
	}

	if requireFirstLogin && !user.IsFirstLogin {
		s.logger.Error("User is not first time login", "phone number", normalizedPhoneNumber)
		return nil, common.ErrUserNotFirstTimeLogin
	}
	if !requireFirstLogin && user.IsFirstLogin {
		s.logger.Error("User is first time login", "phone number", normalizedPhoneNumber)
		return nil, common.ErrUserIsFirstTimeLogin
	}
	return user, nil
}

func (s *authService) createLoginResponse(user users.UserDTO) (*LoginResponse, error) {
	payload := s.generatePayload(user)
	accessToken, refreshToken, err := s.generateTokens(payload)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: user}, nil
}
