package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Kostaaa1/tinylink/internal/data"
)

// func createHashAlias(clientID, url string, length int) string {
// 	s := clientID + url
// 	return fmt.Sprintf("%x", sha1.Sum([]byte(s)))[:length]
// }

type envelope map[string]interface{}

func LogError(r *http.Request, err error) {
	// a.logger.Error(err.Error())
}

func RateLimitExceededResponse(w http.ResponseWriter, r *http.Request, rps float64) {
	// w.Header().Set("Retry-After", fmt.Sprintf("%d"))
	ErrorResponse(w, r, http.StatusTooManyRequests, "Rate limit exceeded, too many requests!")
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, msg interface{}) {
	env := envelope{"error": msg}
	if err := writeJSON(w, status, env, nil); err != nil {
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
	case errors.Is(err, data.ErrURLExists):
		return http.StatusConflict, "Tinylink already exists for this URL"
	case errors.Is(err, data.ErrAliasExists):
		return http.StatusConflict, "Alias not available"
	default:
		return http.StatusInternalServerError, "Internal server error"
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	js = append(js, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func readJSON(r *http.Request, dst interface{}) error {
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	return nil
}
