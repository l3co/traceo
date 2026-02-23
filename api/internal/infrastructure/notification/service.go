package notification

import (
	"context"
	"log/slog"
)

type Service struct {
	email *EmailSender
}

func NewService(email *EmailSender) *Service {
	return &Service{email: email}
}

func (s *Service) NotifySighting(ctx context.Context, userEmail, observation string) error {
	html, err := renderTemplate(sightingEmailTpl, map[string]string{
		"Observation": observation,
	})
	if err != nil {
		return err
	}

	if s.email == nil {
		slog.Warn("email sender not configured, skipping notification",
			"to", userEmail,
		)
		return nil
	}

	return s.email.Send(ctx, userEmail, "Desaparecido foi avistado!", html)
}
