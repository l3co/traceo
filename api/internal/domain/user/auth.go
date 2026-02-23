package user

import "context"

type AuthService interface {
	CreateUser(ctx context.Context, email, password string) (uid string, err error)
	VerifyToken(ctx context.Context, token string) (uid string, err error)
	DeleteUser(ctx context.Context, uid string) error
	ChangePassword(ctx context.Context, uid string, newPassword string) error
	SendPasswordResetEmail(ctx context.Context, email string) error
}
