package errors

import (
	"fmt"
	"net/http"

	"github.com/Kostaaa1/tinylink/internal/interface/utils/jsonutil"
)

type URLExistsError struct {
	Message string
}

func (e URLExistsError) Error() string {
	return e.Message
}

type AliasUsedError struct {
	Message string
}

func (e AliasUsedError) Error() string {
	return e.Message
}

func LogError(r *http.Request, err error) {
	// a.logger.Error(err.Error())
}

func RateLimitExceededResponse(w http.ResponseWriter, r *http.Request, rps float64) {
	w.Header().Set("Retry-After", fmt.Sprintf("%d"))
	ErrorResponse(w, r, http.StatusTooManyRequests, "Rate limit exceeded, too many requests!")
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, msg interface{}) {
	env := jsonutil.Envelope{"error": msg}
	if err := jsonutil.WriteJSON(w, status, env, nil); err != nil {
		// a.logError(r, err)
		w.WriteHeader(500)
	}
}

func ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	// a.logError(r, err)
	ErrorResponse(w, r, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

func FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	ErrorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	ErrorResponse(w, r, http.StatusBadRequest, err.Error())
}

func NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	ErrorResponse(w, r, http.StatusNotFound, "the requested resource could not be found")
}

func MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	ErrorResponse(w, r, http.StatusMethodNotAllowed, "method not allowed for this resource")
}
