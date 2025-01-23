package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Kostaaa1/tinylink/internal/errors"
	"golang.org/x/time/rate"
)

type Ratelimit struct {
	enabled bool
	burst   int
	rps     float64
}

func New(enabled bool, burst int, rps float64) *Ratelimit {
	return &Ratelimit{
		enabled: enabled,
		burst:   burst,
		rps:     rps,
	}
}

func (rl *Ratelimit) Middleware(next http.Handler) http.Handler {
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
		if rl.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				errors.ServerErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			c, ok := clients[ip]
			if !ok {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(rl.rps), rl.burst)}
			}

			c.lastSeen = time.Now()
			if !c.limiter.Allow() {
				mu.Unlock()
				errors.RateLimitExceeded(w, r)
				return
			}

			mu.Unlock()
			next.ServeHTTP(w, r)
		}
	})
}
