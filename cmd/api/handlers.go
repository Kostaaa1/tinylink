package main

import (
	"context"
	"net/http"

	"github.com/Kostaaa1/tinylink/internal/models"
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
		a.errorResponse(w, r, http.StatusBadRequest, "")
		return
	}

	links, err := a.storage.GetAll(ctx, storage.QueryParams{UserID: sessionID})

	if err != nil {
		a.errorResponse(w, r, http.StatusInternalServerError, "failed to get all tinylinks")
		return
	}

	if err := a.writeJSON(w, http.StatusOK, links, nil); err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *app) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL string `json:"url"`
	}

	if err := a.readJSON(r, &input); err != nil {
		a.logger.Info("failed to get cleintID")
		return
	}

	sessionID, _ := getSessionID(r)
	hashKey := generateURLHash(sessionID, input.URL, 8)
	tl := models.Tinylink{
		TinyURL: hashKey,
		URL:     input.URL,
		QR: models.QR{
			ImageURL: "test-image",
			Width:    "450",
			Height:   "450",
		},
	}

	if err := a.storage.Create(ctx, tl, storage.QueryParams{UserID: sessionID}); err != nil {
		a.logError(r, err)
		a.errorResponse(w, r, http.StatusInternalServerError, "failed to create new record")
		return
	}

	if err := a.writeJSON(w, http.StatusOK, envelope{"msg": "data"}, nil); err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *app) DeleteTinylink(w http.ResponseWriter, r *http.Request) {
	sessionID, _ := getSessionID(r)
	tinylink := mux.Vars(r)["tinylink"]

	if err := a.storage.Delete(ctx, storage.QueryParams{UserID: sessionID, ID: tinylink}); err != nil {
		a.logError(r, err)
		a.errorResponse(w, r, http.StatusBadRequest, "failed to delete tinylink")
		return
	}

	if err := a.writeJSON(w, http.StatusOK, envelope{"msg": "tinylink succesfully deleted"}, nil); err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *app) Redirect(w http.ResponseWriter, r *http.Request) {
	sessionID, _ := getSessionID(r)
	tinylink := mux.Vars(r)["tinylink"]

	tl, err := a.storage.Get(ctx, storage.QueryParams{UserID: sessionID, ID: tinylink})
	if err != nil {
		a.logError(r, err)
		a.errorResponse(w, r, http.StatusInternalServerError, "no data under this hash")
		return
	}

	w.Header().Set("Location", tl.URL)
	w.WriteHeader(http.StatusFound)
}
