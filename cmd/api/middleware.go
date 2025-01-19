package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
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

func (a *app) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		clients = make(map[string]*client)
		mu      sync.Mutex
	)
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.cfg.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				a.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			c, ok := clients[ip]
			if !ok {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(a.cfg.limiter.rps), a.cfg.limiter.burst)}
			}

			c.lastSeen = time.Now()
			if !c.limiter.Allow() {
				mu.Unlock()
				a.rateLimitExceeded(w, r)
				return
			}

			mu.Unlock()
			next.ServeHTTP(w, r)
		}
	})
}

func (a *app) persistSessionMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := string(tinylinkSessionKey)
		session, _ := a.cookiestore.Get(r, key)

		if len(session.Values) == 0 {
			fmt.Println("no session in cookie store. Creating...")
			// maybe store other client data? IP, UserAgent, Referer...
			session.Values["session_id"] = generateRandHex(8)
			session.Options.MaxAge = 24 * 3600 // 24 hours
			session.Options.Secure = true      // https only
			session.Options.HttpOnly = true    // prevent javascript access

			if err := session.Save(r, w); err != nil {
				fmt.Println("Failed to save the session?", err)
				a.serverErrorResponse(w, r, err)
				return
			}
		}

		ctx := context.WithValue(r.Context(), tinylinkSessionKey, session)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
