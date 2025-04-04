package handler

import (
	"log/slog"
	"net/http"
)

type ErrorHandler struct {
	logger *slog.Logger
}

func NewErrorHandler(logger *slog.Logger) *ErrorHandler {
	return &ErrorHandler{logger: logger}
}

func (h *ErrorHandler) logError(r *http.Request, err error) {
	h.logger.Error(err.Error())
}

func (h *ErrorHandler) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request, rps float64) {
	h.ErrorResponse(w, r, http.StatusTooManyRequests, "Rate limit exceeded, too many requests!")
}

func (h *ErrorHandler) ErrorResponse(w http.ResponseWriter, r *http.Request, status int, msg interface{}) {
	env := envelope{"error": msg}
	if err := writeJSON(w, status, env, nil); err != nil {
		h.logError(r, err)
		w.WriteHeader(500)
	}
}

func (h *ErrorHandler) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	h.logError(r, err)
	h.ErrorResponse(w, r, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

func (h *ErrorHandler) UnauthorizedResponse(w http.ResponseWriter, r *http.Request) {
	h.ErrorResponse(w, r, http.StatusUnauthorized, "unauthorized request")
}

func (h *ErrorHandler) FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	h.ErrorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (h *ErrorHandler) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	h.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (h *ErrorHandler) InvalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	h.ErrorResponse(w, r, http.StatusUnauthorized, "invalid authentication credentials")
}

func (h *ErrorHandler) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	h.ErrorResponse(w, r, http.StatusNotFound, "the requested resource could not be found")
}

func (h *ErrorHandler) MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	h.ErrorResponse(w, r, http.StatusMethodNotAllowed, "method not allowed for this resource")
}
