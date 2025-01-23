package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func RateLimit(next http.Handler) http.Handler {
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
