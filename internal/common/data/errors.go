package data

import "errors"

var (
	ErrNotFound        = errors.New("record not found")
	ErrRecordExists    = errors.New("record exists")
	ErrUnauthenticated = errors.New("unauthenticated")
)
