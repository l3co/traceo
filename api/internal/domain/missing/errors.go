package missing

import "errors"

var (
	ErrMissingNotFound = errors.New("missing person not found")
	ErrInvalidMissing  = errors.New("invalid missing person data")
)
