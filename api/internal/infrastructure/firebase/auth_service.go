package firebase

import (
	"context"
	"fmt"
	"strings"

	"firebase.google.com/go/v4/auth"

	"github.com/l3co/traceo-api/internal/domain/user"
)

type AuthService struct {
	client *auth.Client
}

func NewAuthService(client *auth.Client) *AuthService {
	return &AuthService{client: client}
}

func (s *AuthService) CreateUser(ctx context.Context, email, password string) (string, error) {
	params := (&auth.UserToCreate{}).
		Email(email).
		Password(password).
		EmailVerified(false)

	record, err := s.client.CreateUser(ctx, params)
	if err != nil {
		if auth.IsEmailAlreadyExists(err) {
			return "", user.ErrEmailAlreadyExists
		}
		return "", fmt.Errorf("firebase auth: creating user: %w", err)
	}

	return record.UID, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, token string) (string, error) {
	token = strings.TrimPrefix(token, "Bearer ")

	decoded, err := s.client.VerifyIDToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("firebase auth: verifying token: %w", err)
	}

	return decoded.UID, nil
}

func (s *AuthService) DeleteUser(ctx context.Context, uid string) error {
	if err := s.client.DeleteUser(ctx, uid); err != nil {
		return fmt.Errorf("firebase auth: deleting user %s: %w", uid, err)
	}
	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, uid string, newPassword string) error {
	params := (&auth.UserToUpdate{}).Password(newPassword)

	if _, err := s.client.UpdateUser(ctx, uid, params); err != nil {
		return fmt.Errorf("firebase auth: changing password for %s: %w", uid, err)
	}
	return nil
}

func (s *AuthService) SendPasswordResetEmail(ctx context.Context, email string) error {
	link, err := s.client.PasswordResetLink(ctx, email)
	if err != nil {
		return fmt.Errorf("firebase auth: generating reset link for %s: %w", email, err)
	}
	_ = link
	return nil
}
