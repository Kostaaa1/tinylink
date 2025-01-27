package session

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"errors"

	myerr "github.com/Kostaaa1/tinylink/internal/errors"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var (
	authKey       = getAuthKey()
	encrtpyionKey = getEncryptionKey()
	cookiestore   = sessions.NewCookieStore(authKey, encrtpyionKey)
)

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

// Middleware for persistence of
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := string(tinylinkSessionKey)
		session, err := cookiestore.Get(r, key)
		if err != nil {
			fmt.Println("error while getting the session: ", err)
		}

		// Do i need to add check if session is valid?
		if len(session.Values) == 0 {
			// maybe store other client data? IP, UserAgent, Referer...
			fmt.Println("no session in cookie store. Creating and saving... Key: ")
			sessionKey := securecookie.GenerateRandomKey(8)
			session.Values["session_id"] = string(sessionKey)
			session.Options.MaxAge = 24 * 3600 // 1 day
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

type contextKey string

const tinylinkSessionKey contextKey = "tinylink_session"

func GetID(r *http.Request) (string, error) {
	session, ok := r.Context().Value(tinylinkSessionKey).(*sessions.Session)
	fmt.Println(session.Options)
	if !ok {
		return "", errors.New("no session found in context")
	}
	s, ok := session.Values["session_id"].(string)
	if ok {
		return s, nil
	}
	return "", errors.New("session is not a string?")
}
