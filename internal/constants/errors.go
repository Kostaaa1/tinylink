package constants

import "errors"

var (
	ErrNotFound        = errors.New("record not found")
	ErrUnauthenticated = errors.New("user is not authenticated")
)
