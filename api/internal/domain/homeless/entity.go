package homeless

import (
	"fmt"
	"strings"
	"time"

	"github.com/l3co/traceo-api/internal/domain/shared"
	"github.com/l3co/traceo-api/pkg/slug"
)

type Homeless struct {
	ID        string
	Name      string
	Nickname  string
	BirthDate time.Time
	Gender    shared.Gender
	Eyes      shared.EyeColor
	Hair      shared.HairColor
	Skin      shared.SkinColor
	PhotoURL  string
	Location  shared.GeoPoint
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (h *Homeless) Age() int {
	if h.BirthDate.IsZero() {
		return 0
	}
	now := time.Now()
	age := now.Year() - h.BirthDate.Year()
	if now.YearDay() < h.BirthDate.YearDay() {
		age--
	}
	return age
}

func (h *Homeless) GenerateSlug() {
	h.Slug = slug.Generate(h.Name, h.ID)
}

func (h *Homeless) NameLowercase() string {
	return strings.ToLower(h.Name)
}

func (h *Homeless) Validate() error {
	if h.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidHomeless)
	}
	if !h.Gender.IsValid() {
		return fmt.Errorf("%w: invalid gender %q", ErrInvalidHomeless, h.Gender)
	}
	if !h.Eyes.IsValid() {
		return fmt.Errorf("%w: invalid eye color %q", ErrInvalidHomeless, h.Eyes)
	}
	if !h.Hair.IsValid() {
		return fmt.Errorf("%w: invalid hair color %q", ErrInvalidHomeless, h.Hair)
	}
	if !h.Skin.IsValid() {
		return fmt.Errorf("%w: invalid skin color %q", ErrInvalidHomeless, h.Skin)
	}
	if !h.BirthDate.IsZero() && h.BirthDate.After(time.Now()) {
		return fmt.Errorf("%w: birth date cannot be in the future", ErrInvalidHomeless)
	}
	return nil
}

// --- Input DTOs ---

type CreateInput struct {
	Name      string
	Nickname  string
	BirthDate time.Time
	Gender    shared.Gender
	Eyes      shared.EyeColor
	Hair      shared.HairColor
	Skin      shared.SkinColor
	PhotoURL  string
	Lat       float64
	Lng       float64
	Address   string
}
