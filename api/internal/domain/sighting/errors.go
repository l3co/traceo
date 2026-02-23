package sighting

import "errors"

var (
	ErrSightingNotFound = errors.New("sighting not found")
	ErrInvalidSighting  = errors.New("invalid sighting")
)
