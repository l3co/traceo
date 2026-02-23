package matching

import (
	"fmt"
	"time"
)

type MatchStatus string

const (
	MatchStatusPending   MatchStatus = "pending"
	MatchStatusConfirmed MatchStatus = "confirmed"
	MatchStatusRejected  MatchStatus = "rejected"
)

func (s MatchStatus) IsValid() bool {
	switch s {
	case MatchStatusPending, MatchStatusConfirmed, MatchStatusRejected:
		return true
	}
	return false
}

type Match struct {
	ID             string
	HomelessID     string
	MissingID      string
	Score          float64
	Status         MatchStatus
	GeminiAnalysis string
	CreatedAt      time.Time
	ReviewedAt     time.Time
}

func (m *Match) Validate() error {
	if m.HomelessID == "" {
		return fmt.Errorf("%w: homeless_id is required", ErrInvalidMatch)
	}
	if m.MissingID == "" {
		return fmt.Errorf("%w: missing_id is required", ErrInvalidMatch)
	}
	if m.Score < 0 || m.Score > 1 {
		return fmt.Errorf("%w: score must be between 0 and 1", ErrInvalidMatch)
	}
	return nil
}
