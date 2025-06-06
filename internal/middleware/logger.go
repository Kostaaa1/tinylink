package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/errorhandler"
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

type MW struct {
	errorhandler.ErrorHandler
}

func (mw MW) Logger(next http.Handler) http.Handler {
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
		slog.Info(msg, attrs...)
	})
}
