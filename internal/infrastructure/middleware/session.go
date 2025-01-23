package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kostaaa1/tinylink/internal/errors"
	"github.com/Kostaaa1/tinylink/internal/interface/utils/sessionutil"
)

func persistSessionMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := string(sessionutil.TinylinkSessionKey)
		session, _ := a.cookiestore.Get(r, key)

		if len(session.Values) == 0 {
			fmt.Println("no session in cookie store. Creating...")
			// maybe store other client data? IP, UserAgent, Referer...
			session.Values["session_id"] = sessionutil.GenerateRandHex(8)
			session.Options.MaxAge = 24 * 3600 // 24 hours
			session.Options.Secure = true      // https only
			session.Options.HttpOnly = true    // prevent javascript access

			if err := session.Save(r, w); err != nil {
				fmt.Println("Failed to save the session?", err)
				errors.ServerErrorResponse(w, r, err)
				return
			}
		}

		ctx := context.WithValue(r.Context(), sessionutil.TinylinkSessionKey, session)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
