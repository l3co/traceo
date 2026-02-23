package user

import (
	"strings"
	"time"
)

type User struct {
	ID            string
	Name          string
	Email         string
	Phone         string
	CellPhone     string
	AvatarURL     string
	AcceptedTerms bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateInput struct {
	Name          string
	Email         string
	Password      string
	Phone         string
	CellPhone     string
	AcceptedTerms bool
}

type UpdateInput struct {
	Name      string
	Phone     string
	CellPhone string
}

func (i *CreateInput) Sanitize() {
	i.Name = strings.TrimSpace(i.Name)
	i.Email = strings.ToLower(strings.TrimSpace(i.Email))
	i.Phone = strings.TrimSpace(i.Phone)
	i.CellPhone = strings.TrimSpace(i.CellPhone)
}

func (i *UpdateInput) Sanitize() {
	i.Name = strings.TrimSpace(i.Name)
	i.Phone = strings.TrimSpace(i.Phone)
	i.CellPhone = strings.TrimSpace(i.CellPhone)
}
