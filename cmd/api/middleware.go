package main

import (
	"context"
	"fmt"
	"net/http"
)

// Recover panic middleware will only occur if panic happens in the same goroutine that executes recoverPanic mdidleware. So if panic occurs in different goroutines (some background processing etc.), those panics will cause app to exit and bring down the server.
func (a *app) recoverPanic(next http.Handler) http.Handler {
	// Creates defered function that will always run in the event of panic.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recover is builtin function that checks if panic occurred
			if err := recover(); err != nil {
				// Set Connection close header that will trigger go http server to close the current connection after response has been sent.
				w.Header().Set("Connection", "close")
				a.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *app) persistSessionMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := a.cookiestore.Get(r, string(tinylinkSessionKey))
		if len(session.Values) == 0 {
			// maybe store other client data? IP, UserAgent, Referer...
			session.Values["session_id"] = createSessionID(16)

			session.Options.MaxAge = 24 * 3600 // 24 hours
			session.Options.Secure = true      // https only
			session.Options.HttpOnly = true    // prevent javascript access

			if err := session.Save(r, w); err != nil {
				a.serverErrorResponse(w, r, err)
				return
			}
		}

		ctx := context.WithValue(r.Context(), tinylinkSessionKey, session)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
