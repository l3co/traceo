package notification

import (
	"context"
	"fmt"
	"log/slog"
)

type Service struct {
	email    *EmailSender
	telegram *TelegramSender
}

func NewService(email *EmailSender, telegram *TelegramSender) *Service {
	return &Service{email: email, telegram: telegram}
}

func (s *Service) NotifyPotentialMatch(ctx context.Context, missingName string, score float64, analysis string) error {
	slog.Info("potential match found",
		"missing_name", missingName,
		"score", score,
	)

	if s.telegram != nil {
		msg := fmt.Sprintf("🔍 *Possível correspondência!*\n*Nome*: _%s_\n*Score*: %.0f%%\n*Análise*: %s",
			missingName, score*100, truncate(analysis, 200))
		if err := s.telegram.SendMessage(ctx, msg); err != nil {
			slog.Error("telegram notification failed", "error", err.Error())
		}
	}

	return nil
}

func (s *Service) NotifyNewHomeless(ctx context.Context, name, birthDate, photoURL, id string) error {
	slog.Info("new homeless registered",
		"name", name,
		"id", id,
	)

	if s.telegram != nil {
		msg := fmt.Sprintf("🆕 *Novo cadastro*\n*Nome*: _%s_\n*Nascimento*: _%s_\n[Ver perfil](https://traceo.me/homeless/%s)",
			name, birthDate, id)
		if err := s.telegram.SendMessage(ctx, msg); err != nil {
			slog.Error("telegram notification failed", "error", err.Error())
		}
	}

	return nil
}

func (s *Service) NotifySighting(ctx context.Context, userEmail, observation string) error {
	html, err := renderTemplate(sightingEmailTpl, map[string]string{
		"Observation": observation,
	})
	if err != nil {
		return err
	}

	if s.email != nil {
		if err := s.email.Send(ctx, userEmail, "Desaparecido foi avistado!", html); err != nil {
			slog.Error("email notification failed", "error", err.Error())
		}
	} else {
		slog.Warn("email sender not configured, skipping email",
			"to", userEmail,
		)
	}

	if s.telegram != nil {
		msg := fmt.Sprintf("👁️ *Novo avistamento*\n*Observação*: %s", truncate(observation, 300))
		if err := s.telegram.SendMessage(ctx, msg); err != nil {
			slog.Error("telegram notification failed", "error", err.Error())
		}
	}

	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
