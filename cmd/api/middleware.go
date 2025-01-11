package main

import (
	"context"
	"log"
	"net/http"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func (a *app) persistSessionMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := a.store.Get(r, string(tinylinkSessionKey))
		if len(session.Values) == 0 {
			session.Values["session_id"] = createSessionID(8)
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
