package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Kostaaa1/tinylink/internal/repository/storage"
	"github.com/gorilla/mux"
)

var (
	ctx = context.Background()
)

func (a *app) Index(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Hello World</title>
	</head>
	<body>
		<h1>Hello, World!</h1>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (a *app) GetAll(w http.ResponseWriter, r *http.Request) {
	sessionID, err := getSessionID(r)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	links, err := a.storage.GetAll(ctx, storage.QueryParams{ClientID: sessionID})
	if err != nil {
		// a.errorResponse(w, r, http.StatusInternalServerError, "failed to get all tinylinks")
		a.serverErrorResponse(w, r, err)
		return
	}

	if err := a.writeJSON(w, http.StatusOK, links, nil); err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *app) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL   string `json:"url"`
		Alias string `json:"alias"`
	}

	if err := a.readJSON(r, &input); err != nil {
		a.logger.Info("failed to get clientID")
		return
	}

	_, err := url.Parse(input.URL)
	if err != nil {
		a.logError(r, err)
		a.errorResponse(w, r, http.StatusBadRequest, "provided URL is invalid")
		return
	}

	sessionID, _ := getSessionID(r)
	qp := storage.QueryParams{ClientID: sessionID}

	var alias string
	if input.Alias == "" {
		alias = createHashAlias(sessionID, input.URL, 8)
	} else {
		alias = input.Alias
		qp.CheckUnique = true
	}

	tl, err := a.newTinylink(input.URL, alias)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	fmt.Println("Alias:", tl)

	newTl, err := a.storage.Create(ctx, tl, qp)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	if err := a.writeJSON(w, http.StatusOK, newTl, nil); err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *app) DeleteTinylink(w http.ResponseWriter, r *http.Request) {
	sessionID, _ := getSessionID(r)
	tinylink := mux.Vars(r)["alias"]

	if err := a.storage.Delete(ctx, storage.QueryParams{ClientID: sessionID, Alias: tinylink}); err != nil {
		a.errorResponse(w, r, http.StatusBadRequest, "failed to delete tinylink")
		return
	}

	if err := a.writeJSON(w, http.StatusOK, envelope{"msg": "tinylink succesfully deleted"}, nil); err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *app) Redirect(w http.ResponseWriter, r *http.Request) {
	sessionID, _ := getSessionID(r)
	tinylink := mux.Vars(r)["alias"]

	tl, err := a.storage.Get(ctx, storage.QueryParams{ClientID: sessionID, Alias: tinylink})
	if err != nil {
		a.errorResponse(w, r, http.StatusInternalServerError, "no data under this hash")
		return
	}

	w.Header().Set("Location", tl.OriginalURL)
	w.WriteHeader(http.StatusFound)
}
