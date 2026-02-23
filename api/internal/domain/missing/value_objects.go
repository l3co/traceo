package missing

import (
	"fmt"

	"github.com/l3co/traceo-api/internal/domain/shared"
)

// Type aliases for backward compatibility within the missing package
type Gender = shared.Gender
type EyeColor = shared.EyeColor
type HairColor = shared.HairColor
type SkinColor = shared.SkinColor
type GeoPoint = shared.GeoPoint

const (
	GenderMale   = shared.GenderMale
	GenderFemale = shared.GenderFemale

	EyeGreen     = shared.EyeGreen
	EyeBlue      = shared.EyeBlue
	EyeBrown     = shared.EyeBrown
	EyeBlack     = shared.EyeBlack
	EyeDarkBrown = shared.EyeDarkBrown

	HairBlack   = shared.HairBlack
	HairBrown   = shared.HairBrown
	HairRedhead = shared.HairRedhead
	HairBlond   = shared.HairBlond

	SkinWhite  = shared.SkinWhite
	SkinBrown  = shared.SkinBrown
	SkinBlack  = shared.SkinBlack
	SkinYellow = shared.SkinYellow
)

// --- Status (Missing-specific) ---

type Status string

const (
	StatusDisappeared Status = "disappeared"
	StatusFound       Status = "found"
)

func (st Status) IsValid() bool {
	switch st {
	case StatusDisappeared, StatusFound:
		return true
	}
	return false
}

func (st Status) Label() string {
	switch st {
	case StatusDisappeared:
		return "Desaparecido"
	case StatusFound:
		return "Encontrado"
	}
	return "Não informado"
}

// --- Validation helpers ---

func validateGender(g Gender) error {
	if !g.IsValid() {
		return fmt.Errorf("%w: invalid gender %q", ErrInvalidMissing, g)
	}
	return nil
}

func validateEyeColor(e EyeColor) error {
	if !e.IsValid() {
		return fmt.Errorf("%w: invalid eye color %q", ErrInvalidMissing, e)
	}
	return nil
}

func validateHairColor(h HairColor) error {
	if !h.IsValid() {
		return fmt.Errorf("%w: invalid hair color %q", ErrInvalidMissing, h)
	}
	return nil
}

func validateSkinColor(s SkinColor) error {
	if !s.IsValid() {
		return fmt.Errorf("%w: invalid skin color %q", ErrInvalidMissing, s)
	}
	return nil
}
