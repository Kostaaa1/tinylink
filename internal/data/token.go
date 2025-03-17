package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"time"
)

type Scope string

var (
	DefaultTokenTTL           = 24 * time.Hour // for authorized users
	ScopeActivation     Scope = "activation"
	ScopeAuthentication Scope = "authentication"
	ScopeTemporary      Scope = "temporary"
	ErrScopeNotValid          = errors.New("scope is not valid")
)

func (s Scope) String() string {
	switch s {
	case ScopeAuthentication:
		return "authentication"
	case ScopeActivation:
		return "activation"
	default:
		return ""
	}
}

func GetScope(v string) Scope {
	switch v {
	case "authentication":
		return ScopeAuthentication
	case "activation":
		return ScopeActivation
	default:
		return ""
	}
}

func (s Scope) IsValid() error {
	switch s {
	case ScopeActivation, ScopeAuthentication:
		return nil
	}
	return ErrScopeNotValid
}

type Token struct {
	PlainText string        `json:"plain_text"`
	Hash      []byte        `json:"-"`
	UserID    string        `json:"-"`
	TTL       time.Duration `json:"-"`
	Expiry    time.Time     `json:"expiry"`
	Scope     Scope         `json:"-"`
}

func GenerateToken(userID string, ttl time.Duration, scope Scope) (*Token, error) {
	token := &Token{
		UserID: userID,
		TTL:    ttl,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randBytes := make([]byte, 16)

	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randBytes)

	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}
