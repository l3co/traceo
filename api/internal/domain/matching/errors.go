package matching

import "errors"

var (
	ErrMatchNotFound = errors.New("match not found")
	ErrInvalidMatch  = errors.New("invalid match")
)
