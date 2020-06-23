package internalerrors

import "errors"

var (
	// ErrNotValid is used when a resource is not valid.
	ErrNotValid = errors.New("not valid")
	// ErrAlreadyExists is used when a resource already exists.
	ErrAlreadyExists = errors.New("already exists")
	// ErrMissing is used when a resource is missing.
	ErrMissing = errors.New("is missing")
)
