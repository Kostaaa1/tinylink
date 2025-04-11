package ratelimiter

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/config"
	"golang.org/x/time/rate"
)

type ratelimit struct {
	enabled bool
	rps     float64
	burst   int
}

func New(conf config.RatelimitConfig) *ratelimit {
	return &ratelimit{
		enabled: conf.Enabled,
		burst:   conf.Burst,
		rps:     conf.RPS,
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
			w.Header().Set("Content-Type", "application/json")

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				// handlers.ServerErrorResponse(w, r, err)
				fmt.Println("error at ratelimit SplitHostPort: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("the server encountered a problem and could not process your request"))
				return
			}

			// testing... [REMOVE IN PROD?]
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
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Rate limit exceeded, too many requests!"))
				return
			}

			next.ServeHTTP(w, r)
		}
	})
}
