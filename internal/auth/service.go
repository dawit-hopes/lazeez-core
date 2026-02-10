package auth

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
	"lazeez-core/internal/key"
)

var (
	accessTokenExpirationMinutes  = 15
	refreshTokenExpirationMinutes = 60 * 24 * 30
)

type AuthService interface {
	CreateUser(ctx context.Context, req UserRequest) error
	GetUserByID(ctx context.Context, id string) (UserDTO, error)
	UpdateUser(ctx context.Context, id string, req UserRequest) error
	DeleteUser(ctx context.Context, id string) error
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
	SetPassword(ctx context.Context, req SetPasswordRequest) (*LoginResponse, error)
	UserLookUp(ctx context.Context, phoneNumber string) (User, error)
	GetAllUsers(ctx context.Context) ([]*UserDTO, error)
	GetUserByBranchID(ctx context.Context, branchID string) (UserDTO, error)
}

type authService struct {
	authRepository   AuthRepository
	branchRepository branch.BranchRepository
	keyService       key.KeyService
	logger           config.Logger
}

func NewAuthService(authRepository AuthRepository, branchRepository branch.BranchRepository, keyService key.KeyService, logger config.Logger) AuthService {
	return &authService{
		authRepository:   authRepository,
		branchRepository: branchRepository,
		keyService:       keyService,
		logger:           logger,
	}
}

func (s *authService) CreateUser(ctx context.Context, req UserRequest) error {
	normalizedPhoneNumber, err := s.validatePhoneNumber(req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to validate phone number", "error", err)
		return err
	}

	// check if user already exists
	if err := s.authRepository.CheckUserExistsByPhoneNumber(ctx, normalizedPhoneNumber); err != nil {
		s.logger.Error("Failed to check user exists by phone number", "error", err)
		return err
	}

	err = s.validateBranch(ctx, req.BranchID, req.MerchantID)
	if err != nil {
		s.logger.Error("Failed to validate branch", "error", err)
		return err
	}

	req.PhoneNumber = normalizedPhoneNumber
	newUser := s.createUserDefaultData(&req)
	err = s.authRepository.CreateUser(ctx, *newUser)
	if err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return err
	}
	return nil
}

func (s *authService) GetUserByID(ctx context.Context, id string) (UserDTO, error) {
	s.logger.Info("Getting user by ID", "id", id)
	user, err := s.authRepository.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user by ID", "error", err)
		return UserDTO{}, err
	}
	result := user.ToDTO()
	return result, nil
}

func (s *authService) UpdateUser(ctx context.Context, id string, req UserRequest) error {

	// check if user already exists
	existingUser, err := s.checkUserExistsByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to check user exists by phone number", "error", err)
		return err
	}
	// User must exist for update
	if existingUser == nil {
		s.logger.Error("User not found", "phone number", req.PhoneNumber)
		return common.ErrUserNotFound
	}

	if req.PhoneNumber != "" {
		normalizedPhoneNumber, err := s.validatePhoneNumber(req.PhoneNumber)
		if err != nil {
			s.logger.Error("Failed to validate phone number", "error", err)
			return err
		}
		req.PhoneNumber = normalizedPhoneNumber
	}

	user := s.updateUserDefaultData(&req, existingUser)
	return s.authRepository.UpdateUser(ctx, *user)
}

func (s *authService) DeleteUser(ctx context.Context, id string) error {
	s.logger.Info("Deleting user", "id", id)
	return s.authRepository.DeleteUser(ctx, id)
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	s.logger.Info("Logging in", "phone number", req.PhoneNumber)
	existingUser, err := s.checkUserExistsByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to check user exists by phone number", "error", err)
		return LoginResponse{}, err
	}
	// User must exist for login
	if existingUser == nil {
		s.logger.Error("User not found", "phone number", req.PhoneNumber)
		return LoginResponse{}, common.ErrUserNotFound
	}
	ok, err := s.keyService.VerifyPassword(req.Password, existingUser.Password)
	if err != nil {
		s.logger.Error("Failed to verify password", "error", err)
		return LoginResponse{}, err
	}
	if !ok {
		s.logger.Error("Invalid password", "phone number", req.PhoneNumber)
		return LoginResponse{}, common.ErrUnAuthorized
	}

	accessToken, refreshToken, err := s.generateTokens(existingUser)
	if err != nil {
		s.logger.Error("Failed to generate tokens", "error", err)
		return LoginResponse{}, err
	}

	return LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         existingUser,
	}, nil
}

func (s *authService) SetPassword(ctx context.Context, req SetPasswordRequest) (*LoginResponse, error) {
	s.logger.Info("Setting password", "phone number", req.PhoneNumber)
	existingUser, err := s.checkUserExistsByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to check user exists by phone number", "error", err)
		return nil, err
	}
	// User must exist to set password
	if existingUser == nil {
		s.logger.Error("User not found", "phone number", req.PhoneNumber)
		return nil, common.ErrUserNotFound
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

	err = s.authRepository.SetPassword(ctx, req.PhoneNumber, encryptedPassword)
	if err != nil {
		s.logger.Error("Failed to set password", "error", err)
		return nil, err
	}

	accessToken, refreshToken, err := s.generateTokens(existingUser)
	if err != nil {
		s.logger.Error("Failed to generate tokens", "error", err)
		return nil, err
	}

	return &LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: existingUser}, nil
}

func (s *authService) UserLookUp(ctx context.Context, phoneNumber string) (User, error) {
	s.logger.Info("Looking up user by phone number", "phone number", phoneNumber)
	return s.authRepository.GetUserByPhoneNumber(ctx, phoneNumber)
}

func (s *authService) GetAllUsers(ctx context.Context) ([]*UserDTO, error) {
	s.logger.Info("Getting all users")
	users, err := s.authRepository.GetAllUsers(ctx)
	if err != nil {
		s.logger.Error("Failed to get all users", "error", err)
		return nil, err
	}
	userDTOs := make([]*UserDTO, len(users))
	for i, user := range users {
		result := user.ToDTO()
		userDTOs[i] = &result
	}
	return userDTOs, nil
}

func (s *authService) GetUserByBranchID(ctx context.Context, branchID string) (UserDTO, error) {
	s.logger.Info("Getting user by branch ID", "branchID", branchID)
	user, err := s.authRepository.GetUserByBranchID(ctx, branchID)
	if err != nil {
		s.logger.Error("Failed to get user by branch ID", "error", err)
		return UserDTO{}, err
	}
	result := user.ToDTO()
	return result, nil
}
