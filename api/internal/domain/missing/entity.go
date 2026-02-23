package missing

import (
	"fmt"
	"strings"
	"time"

	"github.com/l3co/traceo-api/pkg/slug"
)

type Timestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GeoPoint struct {
	Lat float64
	Lng float64
}

type Missing struct {
	ID                  string
	UserID              string
	Name                string
	Nickname            string
	BirthDate           time.Time
	DateOfDisappearance time.Time
	Height              string
	Clothes             string
	Gender              Gender
	Eyes                EyeColor
	Hair                HairColor
	Skin                SkinColor
	PhotoURL            string
	Location            GeoPoint
	Status              Status
	EventReport         string
	TattooDescription   string
	ScarDescription     string
	WasChild            bool
	Slug                string
	NameLowercase       string
	Timestamps
}

func (m *Missing) CalculateWasChild() {
	if m.BirthDate.IsZero() || m.DateOfDisappearance.IsZero() {
		m.WasChild = false
		return
	}
	age := m.DateOfDisappearance.Year() - m.BirthDate.Year()
	if m.DateOfDisappearance.YearDay() < m.BirthDate.YearDay() {
		age--
	}
	m.WasChild = age < 18
}

func (m *Missing) Age() int {
	if m.BirthDate.IsZero() {
		return 0
	}
	now := time.Now()
	age := now.Year() - m.BirthDate.Year()
	if now.YearDay() < m.BirthDate.YearDay() {
		age--
	}
	return age
}

func (m *Missing) GenerateSlug() {
	m.Slug = slug.Generate(m.Name, m.ID)
	m.NameLowercase = strings.ToLower(m.Name)
}

func (m *Missing) HasTattoo() bool {
	return m.TattooDescription != ""
}

func (m *Missing) HasScar() bool {
	return m.ScarDescription != ""
}

func (m *Missing) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidMissing)
	}
	if m.UserID == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidMissing)
	}
	if err := validateGender(m.Gender); err != nil {
		return err
	}
	if err := validateEyeColor(m.Eyes); err != nil {
		return err
	}
	if err := validateHairColor(m.Hair); err != nil {
		return err
	}
	if err := validateSkinColor(m.Skin); err != nil {
		return err
	}
	if !m.DateOfDisappearance.IsZero() && m.DateOfDisappearance.After(time.Now()) {
		return fmt.Errorf("%w: date of disappearance cannot be in the future", ErrInvalidMissing)
	}
	if !m.BirthDate.IsZero() && m.BirthDate.After(time.Now()) {
		return fmt.Errorf("%w: birth date cannot be in the future", ErrInvalidMissing)
	}
	return nil
}

// --- Input DTOs ---

type CreateInput struct {
	UserID              string
	Name                string
	Nickname            string
	BirthDate           time.Time
	DateOfDisappearance time.Time
	Height              string
	Clothes             string
	Gender              Gender
	Eyes                EyeColor
	Hair                HairColor
	Skin                SkinColor
	PhotoURL            string
	Location            GeoPoint
	EventReport         string
	TattooDescription   string
	ScarDescription     string
}

type UpdateInput struct {
	Name                string
	Nickname            string
	BirthDate           time.Time
	DateOfDisappearance time.Time
	Height              string
	Clothes             string
	Gender              Gender
	Eyes                EyeColor
	Hair                HairColor
	Skin                SkinColor
	PhotoURL            string
	Location            GeoPoint
	Status              Status
	EventReport         string
	TattooDescription   string
	ScarDescription     string
}

type ListOptions struct {
	PageSize int
	After    string
	UserID   string
}
