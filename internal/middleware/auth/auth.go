package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Kostaaa1/tinylink/db"
	"github.com/Kostaaa1/tinylink/internal/data"
)

type contextKey string

var (
	SessionKey                     = "tinylink_session"
	userContextKey      contextKey = "user"
	authTokenContextKey contextKey = "auth"
	tempTokenContextKey contextKey = "temp_auth"
)

func AuthTokenFromCtx(ctx context.Context) *data.Token {
	token, _ := ctx.Value(authTokenContextKey).(*data.Token)
	return token
}

func TempTokenFromCtx(ctx context.Context) *data.User {
	token, _ := ctx.Value(tempTokenContextKey).(*data.User)
	return token
}

func UserFromCtx(ctx context.Context) *data.User {
	user, _ := ctx.Value(userContextKey).(*data.User)
	return user
}

func IsAuthenticated(ctx context.Context) bool {
	return UserFromCtx(ctx) != nil
}

func Middleware(tokenStore db.TokenStore, userStore db.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenText string

			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenText = strings.Split(authHeader, "Bearer ")[1]
			}

			if tokenText == "" {
				cookie, err := r.Cookie(SessionKey)
				if err == nil {
					tokenText = cookie.Value
				}
			}

			ctx := r.Context()
			if tokenText != "" {
				token, err := tokenStore.Get(ctx, tokenText)
				if err == nil && token != nil {
					user, err := userStore.GetByID(ctx, token.UserID)
					if err == nil {
						ctx = context.WithValue(ctx, userContextKey, user)
						ctx = context.WithValue(ctx, authTokenContextKey, token)
					}
				}
			}

			if tokenText == "" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/json")
				res := map[string]interface{}{"error": "unauthorized request"}
				b, _ := json.Marshal(res)
				w.Write(b)
				return
			}

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// var (
// 	authKey       = getAuthKey()
// 	encrtpyionKey = getEncryptionKey()
// 	cookiestore   = sessions.NewCookieStore(authKey, encrtpyionKey)
// )

// type contextKey string
// const TinylinkSessionKey contextKey = "tinylink_session"
// func getAuthKey() []byte {
// 	if key := os.Getenv("TINYLINK_AUTH_KEY"); key != "" {
// 		return []byte(key)
// 	}
// 	return securecookie.GenerateRandomKey(32)
// }
// func getEncryptionKey() []byte {
// 	if key := os.Getenv("TINYLINK_ENCRYPTION_KEY"); key != "" {
// 		return []byte(key)
// 	}
// 	return securecookie.GenerateRandomKey(16)
// }
// func Middleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		key := string(TinylinkSessionKey)
// 		session, err := cookiestore.Get(r, key)
// 		if err != nil {
// 			log.Printf("Error retrieving session: %v", err)
// 		}

// 		sessionID, ok := session.Values["session_id"].(string)

// 		if !ok || sessionID == "" {
// 			SessionKey := securecookie.GenerateRandomKey(8)
// 			if SessionKey == nil {
// 				http.Error(w, "Failed to generate session key", http.StatusInternalServerError)
// 				return
// 			}
// 			session.Values["session_id"] = string(SessionKey)

// 			session.Options = &sessions.Options{
// 				Path:     "/",
// 				MaxAge:   24 * 3600,
// 				HttpOnly: true,
// 				Secure:   r.TLS != nil,
// 			}

// 			if err := session.Save(r, w); err != nil {
// 				log.Printf("Error saving session: %v", err)
// 				http.Error(w, "Failed to save session", http.StatusInternalServerError)
// 				return
// 			}
// 		}

// 		ctx := context.WithValue(r.Context(), TinylinkSessionKey, session)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

// func GetID(r *http.Request) (string, error) {
// 	session, ok := r.Context().Value(TinylinkSessionKey).(*sessions.Session)
// 	if !ok {
// 		return "", errors.New("no session found in context")
// 	}
// 	s, ok := session.Values["session_id"].(string)
// 	if ok {
// 		return s, nil
// 	}
// 	return "", errors.New("session is not a string?")
// }
