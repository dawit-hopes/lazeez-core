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
	CheckUserExistsByPhoneNumber(ctx context.Context, phoneNumber string) error
	GetAllUsers(ctx context.Context) ([]*User, error)
	GetUserByBranchID(ctx context.Context, branchID string) (User, error)
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
	filter := map[string]any{"id": id, "is_deleted": false}
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
	filter := map[string]any{"phone_number": phoneNumber, "is_deleted": false}
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
	r.logger.Info("Updating user and ID is", "user", user, "id", user.ID)

	filter := map[string]any{"id": user.ID}
	_, err := r.dal.Update(ctx, filter, &user)
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
	existingUser, err := r.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		r.logger.Error("user not found", "error", err)
		return err
	}

	user := User{Password: password}
	filter := map[string]any{"id": existingUser.ID}
	_, err = r.dal.Update(ctx, filter, &user)
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

func (r *authRepository) CheckUserExistsByPhoneNumber(ctx context.Context, phoneNumber string) error {
	filter := map[string]any{"phone_number": phoneNumber, "is_deleted": false}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		r.logger.Error("failed to check user exists by phone number", "error", err)
		return err
	}
	return common.ErrUserWithInformationAlreadyExists
}

func (r *authRepository) GetAllUsers(ctx context.Context) ([]*User, error) {
	results, err := r.dal.List(ctx, map[string]any{
		"is_deleted": false,
	}, 0, 0)
	if err != nil {
		r.logger.Error("failed to get all users", "error", err)
		return nil, err
	}
	return results, nil
}

func (r *authRepository) GetUserByBranchID(ctx context.Context, branchID string) (User, error) {
	filter := map[string]any{"branch_id": branchID, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("user not found", "error", err)
			return User{}, common.ErrUserNotFound
		}
		r.logger.Error("failed to get user by branch ID", "error", err)
		return User{}, err
	}
	return *result, nil
}
