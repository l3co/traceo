package notification

import "context"

type Notifier interface {
	NotifySighting(ctx context.Context, userEmail, observation string) error
	NotifyNewHomeless(ctx context.Context, name, birthDate, photoURL, id string) error
}
