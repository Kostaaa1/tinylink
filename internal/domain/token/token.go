package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"time"
)

type Scope string

var (
	ScopeActivation     Scope = "activation"
	ScopeAuthentication Scope = "authentication"
	ScopeTemporary      Scope = "anonymous_session"
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

func GenerateToken(userID string) *Token {
	scope, ttl := determineTokenScope(userID)

	token := &Token{
		UserID: userID,
		TTL:    ttl,
		Scope:  scope,
	}

	token.Expiry = time.Now().Add(ttl).Truncate(time.Second)
	token.PlainText = generateRandomToken(16)
	token.Hash = hashIt(token.PlainText)

	return token
}

func determineTokenScope(userID string) (Scope, time.Duration) {
	if userID == "" {
		return ScopeTemporary, time.Hour * 3
	}
	return ScopeAuthentication, time.Hour * 24
}

func hashIt(v string) []byte {
	hash := sha256.Sum256([]byte(v))
	return hash[:]
}

func generateRandomToken(n int) string {
	randBytes := make([]byte, n)
	rand.Read(randBytes)
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randBytes)
}
