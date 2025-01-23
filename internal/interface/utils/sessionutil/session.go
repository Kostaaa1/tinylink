package sessionutil

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey string

const TinylinkSessionKey contextKey = "tinylink_session"

func GetID(r *http.Request) (string, error) {
	session, ok := r.Context().Value(TinylinkSessionKey).(*sessions.Session)
	if !ok {
		return "", errors.New("no session found in context")
	}
	s, ok := session.Values["session_id"].(string)
	if ok {
		return s, nil
	}
	return "", errors.New("session is not a string?")
}

func GenerateRandHex(l int) string {
	b := make([]byte, l)
	rand.Read(b)
	return hex.EncodeToString(b)
}
