package handler

import "time"

// --- Request DTOs ---

type CreateUserRequest struct {
	Name          string `json:"name" validate:"required,max=150"`
	Email         string `json:"email" validate:"required,email"`
	Password      string `json:"password" validate:"required,min=6,max=100"`
	Phone         string `json:"phone,omitempty" validate:"omitempty,max=20"`
	CellPhone     string `json:"cell_phone,omitempty" validate:"omitempty,max=20"`
	AcceptedTerms bool   `json:"accepted_terms" validate:"required"`
}

type UpdateUserRequest struct {
	Name      string `json:"name" validate:"required,max=150"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,max=20"`
	CellPhone string `json:"cell_phone,omitempty" validate:"omitempty,max=20"`
}

type ChangePasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=6,max=100"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// --- Response DTOs ---

type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	CellPhone string `json:"cell_phone,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toUserResponse(id, name, email, phone, cellPhone, avatarURL string, createdAt, updatedAt time.Time) UserResponse {
	return UserResponse{
		ID:        id,
		Name:      name,
		Email:     email,
		Phone:     phone,
		CellPhone: cellPhone,
		AvatarURL: avatarURL,
		CreatedAt: createdAt.Format(time.RFC3339),
		UpdatedAt: updatedAt.Format(time.RFC3339),
	}
}
