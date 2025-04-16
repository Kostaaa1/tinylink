package middleware

import (
	"net/http"
)

// Recover panic middleware will only occur if panic happens in the same goroutine that executes recoverPanic mdidleware. So if panic occurs in different goroutines (some background processing etc.), those panics will cause app to exit and bring down the server.
func (mw MW) RecoverPanic(next http.Handler) http.Handler {
	// Creates defered function that will always run in the event of panic.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recover is builtin function that checks if panic occurred
			if err := recover(); err != nil {
				// Set Connection close header that will trigger go http server to close the current connection after response has been sent.
				w.Header().Set("Connection", "close")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("the server encountered a problem and could not process your request"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
