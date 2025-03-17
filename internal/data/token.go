package data

import (
	"errors"
	"time"
)

type Scope string

var (
	ScopeActivation     Scope = "activation"
	ScopeAuthentication Scope = "authentication"
	ErrScopeNotValid          = errors.New("scope is not valid")
)

func (s Scope) IsValid() error {
	switch s {
	case ScopeActivation, ScopeAuthentication:
		return nil
	}
	return ErrScopeNotValid
}

type Token struct {
	PlainText string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     Scope
}
