package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Kostaaa1/tinylink/api/utils/jsonutil"
)

var (
	ErrAliasExists = errors.New("this alias is not available. All aliasses must be unique")
	ErrURLExists   = errors.New("you've already created a tinylink with this URL")
)

func LogError(r *http.Request, err error) {
	// a.logger.Error(err.Error())
}

func RateLimitExceededResponse(w http.ResponseWriter, r *http.Request, rps float64) {
	// w.Header().Set("Retry-After", fmt.Sprintf("%d"))
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
	fmt.Println("Error: ", err)
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

func MapErrorToStatus(err error) (int, string) {
	switch {
	case errors.Is(err, ErrURLExists):
		return http.StatusConflict, "Tinylink already exists for this URL"
	case errors.Is(err, ErrAliasExists):
		return http.StatusConflict, "Alias not available"
	default:
		return http.StatusInternalServerError, "Internal server error"
	}
}
