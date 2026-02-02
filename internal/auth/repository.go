package auth

import "context"

type AuthRepository interface {
	CreateUser(ctx context.Context, user User) error
	GetUserByID(ctx context.Context, id string) (User, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (User, error)
	UpdateUser(ctx context.Context, user User) error
	SetPassword(ctx context.Context, phoneNumber string, password string) error
	DeleteUser(ctx context.Context, id string) error
}
