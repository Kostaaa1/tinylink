package main

import "net/http"

func (a *app) logError(r *http.Request, err error) {
	a.logger.Error(err.Error())
}

func (a *app) errorResponse(w http.ResponseWriter, r *http.Request, status int, msg string) {
	env := envelope{
		"error": msg,
	}
	if err := a.writeJSON(w, status, env, nil); err != nil {
		a.logError(r, err)
		w.WriteHeader(500)
	}
}

func (a *app) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	a.logError(r, err)
	a.errorResponse(w, r, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

func (a *app) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	a.errorResponse(w, r, http.StatusNotFound, "the requested resource could not be found")
}

func (a *app) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	a.errorResponse(w, r, http.StatusMethodNotAllowed, "method not allowed for this resource")
}
