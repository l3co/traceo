package homeless

import "errors"

var (
	ErrHomelessNotFound = errors.New("homeless not found")
	ErrInvalidHomeless  = errors.New("invalid homeless")
)
