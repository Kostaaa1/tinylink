package ratelimiter

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Kostaaa1/tinylink/internal/errors"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"golang.org/x/time/rate"
)

type ratelimit struct {
	enabled bool
	rps     float64
	burst   int
}

func New(cfg config.RatelimitConfig) *ratelimit {
	return &ratelimit{
		enabled: cfg.Enabled,
		burst:   cfg.Burst,
		rps:     cfg.RPS,
	}
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*client)
)

func (rl *ratelimit) Middleware(next http.Handler) http.Handler {
	// Launch a background goroutine to clean up old clients.
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

			// testing...
			if ip == "::1" {
				ip = "127.0.0.1"
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(rl.rps), rl.burst),
				}
			}
			clients[ip].lastSeen = time.Now()

			mu.Unlock()

			if !clients[ip].limiter.Allow() {
				errors.RateLimitExceededResponse(w, r, rl.rps)
				return
			}

			next.ServeHTTP(w, r)
		}
	})
}
