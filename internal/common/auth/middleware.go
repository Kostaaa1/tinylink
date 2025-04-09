package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func unauthorizeResponse(w http.ResponseWriter, r *http.Request) {
	d := map[string]string{"error": "unauthorized request"}
	b, _ := json.Marshal(d)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(b)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		token := strings.TrimPrefix(bearer, "Bearer ")

		if token == "" {
			cookie, err := r.Cookie(SessionKey)
			if err != nil || cookie.Value == "" {
				unauthorizeResponse(w, r)
				return
			}
			token = cookie.Value
		}

		claims, err := VerifyJWT(token)
		if err != nil {
			fmt.Println("Token verification failed: ", err)
			unauthorizeResponse(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
