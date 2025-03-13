package session

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var (
	authKey       = getAuthKey()
	encrtpyionKey = getEncryptionKey()
	cookiestore   = sessions.NewCookieStore(authKey, encrtpyionKey)
)

type contextKey string

const tinylinkSessionKey contextKey = "tinylink_session"

func getAuthKey() []byte {
	if key := os.Getenv("TINYLINK_AUTH_KEY"); key != "" {
		return []byte(key)
	}
	return securecookie.GenerateRandomKey(32)
}

func getEncryptionKey() []byte {
	if key := os.Getenv("TINYLINK_ENCRYPTION_KEY"); key != "" {
		return []byte(key)
	}
	return securecookie.GenerateRandomKey(16)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := string(tinylinkSessionKey)
		session, err := cookiestore.Get(r, key)
		if err != nil {
			log.Printf("Error retrieving session: %v", err)
		}

		sessionID, ok := session.Values["session_id"].(string)
		if !ok || sessionID == "" {
			sessionKey := securecookie.GenerateRandomKey(8)
			if sessionKey == nil {
				http.Error(w, "Failed to generate session key", http.StatusInternalServerError)
				return
			}
			session.Values["session_id"] = string(sessionKey)

			session.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   24 * 3600,
				HttpOnly: true,
				Secure:   r.TLS != nil,
			}

			if err := session.Save(r, w); err != nil {
				log.Printf("Error saving session: %v", err)
				http.Error(w, "Failed to save session", http.StatusInternalServerError)
				return
			}
		}

		ctx := context.WithValue(r.Context(), tinylinkSessionKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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
