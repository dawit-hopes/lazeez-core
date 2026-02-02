package auth

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
)

type AuthService interface {
	CreateUser(ctx context.Context, req UserRequest) error
	GetUserByID(ctx context.Context, id string) (User, error)
	UpdateUser(ctx context.Context, req UserRequest) error
	DeleteUser(ctx context.Context, id string) error
}

type authService struct {
	authRepository   AuthRepository
	branchRepository branch.BranchRepository
	logger           config.Logger
}

func NewAuthService(authRepository AuthRepository, branchRepository branch.BranchRepository, logger config.Logger) AuthService {
	return &authService{
		authRepository:   authRepository,
		branchRepository: branchRepository,
		logger:           logger,
	}
}

func (s *authService) CreateUser(ctx context.Context, req UserRequest) error {
	s.logger.Info("Creating user by phone number", "phone number", req.PhoneNumber)
	// check if user already exists
	_, err := s.checkUserExistsByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to check user exists", "error", err)
		return err
	}
	err = s.validateBranch(ctx, req.BranchID, req.MerchantID)
	if err != nil {
		s.logger.Error("Failed to validate branch", "error", err)
		return err
	}

	normalizedPhoneNumber, err := s.validatePhoneNumber(req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to validate phone number", "error", err)
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

func (s *authService) GetUserByID(ctx context.Context, id string) (User, error) {
	s.logger.Info("Getting user by ID", "id", id)
	return s.authRepository.GetUserByID(ctx, id)
}

func (s *authService) UpdateUser(ctx context.Context, req UserRequest) error {

	// check if user already exists
	existingUser, err := s.checkUserExistsByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to check user exists by phone number", "error", err)
		return err
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
