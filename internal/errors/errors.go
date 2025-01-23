package errors

import (
	"net/http"

	"github.com/Kostaaa1/tinylink/internal/interface/utils/jsonutil"
)

func LogError(r *http.Request, err error) {
	// a.logger.Error(err.Error())
}

func RateLimitExceeded(w http.ResponseWriter, r *http.Request) {
	ErrorResponse(w, r, http.StatusTooManyRequests, "Rate limit exceeded, too many requests!")
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, msg string) {
	env := jsonutil.Envelope{
		"error": msg,
	}
	if err := jsonutil.WriteJSON(w, status, env, nil); err != nil {
		// a.logError(r, err)
		w.WriteHeader(500)
	}
}

func ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	// a.logError(r, err)
	ErrorResponse(w, r, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

func NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	ErrorResponse(w, r, http.StatusNotFound, "the requested resource could not be found")
}

func MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	ErrorResponse(w, r, http.StatusMethodNotAllowed, "method not allowed for this resource")
}
