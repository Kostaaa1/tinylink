package data

import "errors"

var (
	ErrNotFound      = errors.New("record not found")
	ErrRecordExists  = errors.New("record exists")
	ErrNonAuthorized = errors.New("missing user id (from bearer token) AND session id (from cookie)")
)
