package sighting

import (
	"fmt"
	"time"
)

type GeoPoint struct {
	Lat float64
	Lng float64
}

type Sighting struct {
	ID          string
	MissingID   string
	Location    GeoPoint
	Observation string
	CreatedAt   time.Time
}

func (s *Sighting) Validate() error {
	if s.MissingID == "" {
		return fmt.Errorf("%w: missing_id is required", ErrInvalidSighting)
	}
	if s.Location.Lat == 0 && s.Location.Lng == 0 {
		return fmt.Errorf("%w: location is required", ErrInvalidSighting)
	}
	if s.Observation == "" {
		return fmt.Errorf("%w: observation is required", ErrInvalidSighting)
	}
	return nil
}
