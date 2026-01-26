package auth

import "context"

type AuthRepository interface {
	CreateUser(ctx context.Context, user User) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (User, error)
	GetUserByUserName(ctx context.Context, userName string) (User, error)
	UpdateUser(ctx context.Context, user User) (User, error)
	DeleteUser(ctx context.Context, id string) error
}
