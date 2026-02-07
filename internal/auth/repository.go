package auth

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user User) error
	GetUserByID(ctx context.Context, id string) (User, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (User, error)
	UpdateUser(ctx context.Context, user User) error
	SetPassword(ctx context.Context, phoneNumber string, password string) error
	DeleteUser(ctx context.Context, id string) error
}

type authRepository struct {
	dal    *common.DAL[*User]
	logger config.Logger
}

func NewAuthRepository(dal *common.DAL[*User], logger config.Logger) AuthRepository {
	return &authRepository{dal: dal, logger: logger}
}

func (r *authRepository) CreateUser(ctx context.Context, user User) error {
	_, err := r.dal.Create(ctx, &user)
	if err != nil {
		r.logger.Error("failed to create user", "error", err)
		return err
	}
	return nil
}

func (r *authRepository) GetUserByID(ctx context.Context, id string) (User, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return User{}, common.ErrUserNotFound
		}
		r.logger.Error("failed to get user by ID", "error", err)
		return User{}, err
	}
	return *result, nil
}

func (r *authRepository) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (User, error) {
	filter := map[string]any{"phone_number": phoneNumber}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return User{}, common.ErrUserNotFound
		}
		r.logger.Error("failed to get user by phone number", "error", err)
		return User{}, err
	}
	return *result, nil
}

func (r *authRepository) UpdateUser(ctx context.Context, user User) error {
	_, err := r.dal.Update(ctx, user.ID, &user)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return common.ErrUserNotFound
		}
		r.logger.Error("failed to update user", "error", err)
		return err
	}
	return nil
}

func (r *authRepository) DeleteUser(ctx context.Context, id string) error {
	err := r.dal.Delete(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return common.ErrUserNotFound
		}
		r.logger.Error("failed to delete user", "error", err)
		return err
	}
	return nil
}

func (r *authRepository) SetPassword(ctx context.Context, phoneNumber string, password string) error {
	// First, fetch the user by phone number to get the ID
	existingUser, err := r.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		r.logger.Error("user not found", "error", err)
		return err
	}

	// Update the user's password using the user's ID
	user := User{Password: password}
	_, err = r.dal.Update(ctx, existingUser.ID, &user)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return common.ErrUserNotFound
		}
		r.logger.Error("failed to set password", "error", err)
		return err
	}
	return nil
}
