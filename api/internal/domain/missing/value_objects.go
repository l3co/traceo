package missing

import "fmt"

// --- Gender ---

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

func (g Gender) IsValid() bool {
	switch g {
	case GenderMale, GenderFemale:
		return true
	}
	return false
}

func (g Gender) Label() string {
	switch g {
	case GenderMale:
		return "Masculino"
	case GenderFemale:
		return "Feminino"
	}
	return "Não informado"
}

// --- EyeColor ---

type EyeColor string

const (
	EyeGreen     EyeColor = "green"
	EyeBlue      EyeColor = "blue"
	EyeBrown     EyeColor = "brown"
	EyeBlack     EyeColor = "black"
	EyeDarkBrown EyeColor = "dark_brown"
)

func (e EyeColor) IsValid() bool {
	switch e {
	case EyeGreen, EyeBlue, EyeBrown, EyeBlack, EyeDarkBrown:
		return true
	}
	return false
}

func (e EyeColor) Label() string {
	switch e {
	case EyeGreen:
		return "Verde"
	case EyeBlue:
		return "Azul"
	case EyeBrown:
		return "Castanho"
	case EyeBlack:
		return "Pretos"
	case EyeDarkBrown:
		return "Castanho Escuro"
	}
	return "Não informado"
}

// --- HairColor ---

type HairColor string

const (
	HairBlack   HairColor = "black"
	HairBrown   HairColor = "brown"
	HairRedhead HairColor = "redhead"
	HairBlond   HairColor = "blond"
)

func (h HairColor) IsValid() bool {
	switch h {
	case HairBlack, HairBrown, HairRedhead, HairBlond:
		return true
	}
	return false
}

func (h HairColor) Label() string {
	switch h {
	case HairBlack:
		return "Preto"
	case HairBrown:
		return "Castanho"
	case HairRedhead:
		return "Ruivo"
	case HairBlond:
		return "Loiro"
	}
	return "Não informado"
}

// --- SkinColor ---

type SkinColor string

const (
	SkinWhite  SkinColor = "white"
	SkinBrown  SkinColor = "brown"
	SkinBlack  SkinColor = "black"
	SkinYellow SkinColor = "yellow"
)

func (s SkinColor) IsValid() bool {
	switch s {
	case SkinWhite, SkinBrown, SkinBlack, SkinYellow:
		return true
	}
	return false
}

func (s SkinColor) Label() string {
	switch s {
	case SkinWhite:
		return "Branca"
	case SkinBrown:
		return "Parda"
	case SkinBlack:
		return "Negra"
	case SkinYellow:
		return "Amarela"
	}
	return "Não informado"
}

// --- Status ---

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
