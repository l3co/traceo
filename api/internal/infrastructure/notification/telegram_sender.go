package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type TelegramSender struct {
	botToken string
	chatID   string
	client   *http.Client
}

func NewTelegramSender(botToken, chatID string) *TelegramSender {
	return &TelegramSender{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *TelegramSender) SendMessage(ctx context.Context, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	payload := map[string]string{
		"chat_id":    t.chatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling telegram payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		slog.Error("telegram api error",
			"status", resp.StatusCode,
		)
		return fmt.Errorf("telegram api returned status %d", resp.StatusCode)
	}

	return nil
}
