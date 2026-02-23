package user

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	repo Repository
	auth AuthService
}

func NewService(repo Repository, auth AuthService) *Service {
	return &Service{repo: repo, auth: auth}
}

func (s *Service) Create(ctx context.Context, input *CreateInput) (*User, error) {
	input.Sanitize()

	if input.Name == "" || input.Email == "" || input.Password == "" {
		return nil, fmt.Errorf("%w: name, email, and password are required", ErrInvalidInput)
	}

	if !input.AcceptedTerms {
		return nil, ErrTermsNotAccepted
	}

	existing, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil && err != ErrUserNotFound {
		return nil, fmt.Errorf("checking existing email: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	uid, err := s.auth.CreateUser(ctx, input.Email, input.Password)
	if err != nil {
		return nil, fmt.Errorf("creating auth user: %w", err)
	}

	now := time.Now()
	user := &User{
		ID:            uid,
		Name:          input.Name,
		Email:         input.Email,
		Phone:         input.Phone,
		CellPhone:     input.CellPhone,
		AcceptedTerms: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		_ = s.auth.DeleteUser(ctx, uid)
		return nil, fmt.Errorf("saving user: %w", err)
	}

	return user, nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id string, input *UpdateInput) (*User, error) {
	input.Sanitize()

	if id == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Name = input.Name
	user.Phone = input.Phone
	user.CellPhone = input.CellPhone
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}

	return user, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidInput)
	}

	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	if err := s.auth.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("deleting auth user: %w", err)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) ChangePassword(ctx context.Context, userID, newPassword string) error {
	if userID == "" || newPassword == "" {
		return fmt.Errorf("%w: user id and new password are required", ErrInvalidInput)
	}
	return s.auth.ChangePassword(ctx, userID, newPassword)
}

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	if email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	return s.auth.SendPasswordResetEmail(ctx, email)
}
