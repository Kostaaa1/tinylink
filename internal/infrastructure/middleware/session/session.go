package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"errors"

	myerr "github.com/Kostaaa1/tinylink/internal/errors"
	"github.com/gorilla/sessions"
)

type Session struct {
	cookiestore *sessions.CookieStore
}

func New(c *sessions.CookieStore) *Session {
	return &Session{
		cookiestore: c,
	}
}

type contextKey string

const tinylinkSessionKey contextKey = "tinylink_session"

func GetID(r *http.Request) (string, error) {
	session, ok := r.Context().Value(tinylinkSessionKey).(*sessions.Session)
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

func (s *Session) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := string(tinylinkSessionKey)
		session, _ := s.cookiestore.Get(r, key)

		// Do i need to add check if session is valid?

		if len(session.Values) == 0 {
			fmt.Println("no session in cookie store. Creating...")
			// maybe store other client data? IP, UserAgent, Referer...
			session.Values["session_id"] = GenerateRandHex(8)
			session.Options.MaxAge = 24 * 3600 // 24 hours
			session.Options.Secure = true      // https only
			session.Options.HttpOnly = true    // prevent javascript access

			if err := session.Save(r, w); err != nil {
				fmt.Println("Failed to save the session?", err)
				myerr.ServerErrorResponse(w, r, err)
				return
			}
		}

		ctx := context.WithValue(r.Context(), tinylinkSessionKey, session)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
