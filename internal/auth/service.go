package auth

import (
	"context"
	"lazeez-core/config"
	"time"

	"github.com/google/uuid"
)

type AuthService interface {
	CreateUser(ctx context.Context, user User) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
	GetUserByUserName(ctx context.Context, userName string) (User, error)
	UpdateUser(ctx context.Context, userName string, user User) (User, error)
	DeleteUser(ctx context.Context, id string) error
}

type authService struct {
	authRepository AuthRepository
	logger         config.Logger
}

func NewAuthService(authRepository AuthRepository, logger config.Logger) AuthService {
	return &authService{
		authRepository: authRepository,
		logger:         logger,
	}
}

func (s *authService) CreateUser(ctx context.Context, user User) (User, error) {
	s.logger.Info("Creating user", "user", user)
	newUser := s.createUserDefaultData(&user)
	createdUser, err := s.authRepository.CreateUser(ctx, *newUser)
	if err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return createdUser, err
	}
	return createdUser, nil
}

func (s *authService) GetUserByID(ctx context.Context, id string) (User, error) {
	s.logger.Info("Getting user by ID", "id", id)
	return s.authRepository.GetUserByID(ctx, id)
}

func (s *authService) GetUserByUserName(ctx context.Context, userName string) (User, error) {
	s.logger.Info("Getting user by user name", "user name", userName)
	return s.authRepository.GetUserByUserName(ctx, userName)
}

func (s *authService) UpdateUser(ctx context.Context, userName string, user User) (User, error) {
	s.logger.Info("Updating user", "user name", userName)
	existingUser, err := s.GetUserByUserName(ctx, userName)
	if err != nil {
		s.logger.Error("Failed to get user by user name", "error", err)
		return existingUser, err
	}
	return s.authRepository.UpdateUser(ctx, user)
}

func (s *authService) DeleteUser(ctx context.Context, id string) error {
	s.logger.Info("Deleting user", "id", id)
	return s.authRepository.DeleteUser(ctx, id)
}

func (s *authService) createUserDefaultData(user *User) *User {
	now := time.Now()
	user.ID = uuid.New().String()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsFirstLogin = true
	user.IsLocked = false
	user.LoggingAttempts = 0
	user.Role = RoleBranchManager
	return user
}
