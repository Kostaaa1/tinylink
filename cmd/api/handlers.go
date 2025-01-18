package main

import (
	"context"
	"errors"
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
	clientID, err := createClientID(r)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	links, err := a.storage.GetAll(ctx, storage.QueryParams{ClientID: clientID})
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	if err := a.writeJSON(w, http.StatusOK, links, nil); err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *app) validateCreateInput(clientID, URL string, alias *string) error {
	_, err := url.Parse(URL)
	if err != nil {
		return errors.New("format of the provided URL is not valid")
	}
	if err := a.storage.ValidateOriginalURL(ctx, clientID, URL); err != nil {
		return err
	}

	if *alias == "" {
		*alias = createHashAlias(clientID, URL, 8)
	} else {
		if err := a.storage.ValidateAlias(ctx, *alias); err != nil {
			return err
		}
	}

	return nil
}

func (a *app) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL   string `json:"url"`
		Alias string `json:"alias"`
	}
	if err := a.readJSON(r, &input); err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	clientID, err := createClientID(r)
	if err != nil {
		a.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.validateCreateInput(clientID, input.URL, &input.Alias); err != nil {
		a.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	tl, err := a.newTinylink(input.URL, input.Alias)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	qp := storage.QueryParams{ClientID: clientID, Alias: input.Alias}
	if err := a.storage.Create(ctx, tl, qp); err != nil {
		a.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.writeJSON(w, http.StatusOK, tl, nil); err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *app) DeleteTinylink(w http.ResponseWriter, r *http.Request) {
	clientID, _ := createClientID(r)
	tinylink := mux.Vars(r)["alias"]

	if err := a.storage.Delete(ctx, storage.QueryParams{ClientID: clientID, Alias: tinylink}); err != nil {
		a.errorResponse(w, r, http.StatusBadRequest, "failed to delete tinylink")
		return
	}

	if err := a.writeJSON(w, http.StatusOK, envelope{"msg": "tinylink succesfully deleted"}, nil); err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *app) Redirect(w http.ResponseWriter, r *http.Request) {
	clientID, err := createClientID(r)
	if err != nil {
		a.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	tinylink := mux.Vars(r)["alias"]

	tl, err := a.storage.Get(ctx, storage.QueryParams{ClientID: clientID, Alias: tinylink})
	if err != nil {
		a.errorResponse(w, r, http.StatusInternalServerError, "no data under this hash")
		return
	}

	w.Header().Set("Location", tl.OriginalURL)
	w.WriteHeader(http.StatusFound)
}
