package loggermw

import (
	"log/slog"
	"net/http"
	"time"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusResponseWriter) GetStatus() int {
	return w.statusCode
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.RequestURI()

		wrappedWriter := &statusResponseWriter{ResponseWriter: w}
		next.ServeHTTP(wrappedWriter, r)

		statusCode := wrappedWriter.GetStatus()

		attrs := []any{
			slog.Duration("latency", time.Since(start)),
			slog.String("method", r.Method),
			slog.Int("status", statusCode),
			slog.String("path", path),
		}

		msg := "request"
		if statusCode >= 500 {
			slog.Error(msg, attrs...)
			return
		}
		slog.Debug(msg, attrs...)
	})
}
