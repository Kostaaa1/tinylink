package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
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
}

func (a *app) isURLInStore(ctx context.Context, key, inputURL string) bool {
	urls, _ := a.rdb.HVals(ctx, key).Result()
	for _, u := range urls {
		if u == inputURL {
			return true
		}
	}
	return false
}

func (a *app) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL    string `json:"url"`
		Domain string `json:"domain"`
		Alias  string `json:"alias"`
	}

	if err := a.readJSON(r, &input); err != nil {
		a.logger.Info("failed to get cleintID")
		return
	}

	sessionID, err := getSessionID(r)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	hashKey := generateURLHash(sessionID, input.URL)[:8]
	redisKey := fmt.Sprintf("client:%s:tinylink:%s", sessionID, hashKey)

	qr := map[string]interface{}{
		"url":          input.URL,
		"domain":       input.Domain,
		"alias":        input.Alias,
		"qr:image_url": "test",
		"qr:width":     "test",
		"qr:height":    "test",
	}

	ctx := context.Background()
	if err := a.rdb.HSet(ctx, redisKey, qr).Err(); err != nil {
		a.logError(r, err)
		panic(err)
	}
}

func (a *app) DeleteTinylink(w http.ResponseWriter, r *http.Request) {
	sessionID, err := getSessionID(r)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	redisKey := fmt.Sprintf("client:%s:data", sessionID)

	ctx := context.Background()
	if err := a.rdb.Del(ctx, redisKey).Err(); err != nil {
		a.errorResponse(w, r, http.StatusBadRequest, "failed to delete tinylink")
		return
	}

	if err := a.writeJSON(w, http.StatusOK, envelope{"msg": "tinylink succesfully deleted"}, nil); err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *app) Redirect(w http.ResponseWriter, r *http.Request) {
	sessionID, err := getSessionID(r)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	tinylink := mux.Vars(r)["tinylink"]
	redisKey := fmt.Sprintf("client:%s:data", sessionID)

	ctx := context.Background()
	val, err := a.rdb.HGet(ctx, redisKey, tinylink).Result()
	if err != nil {
		http.Error(w, "Failed to get the url by key", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", val)
	w.WriteHeader(http.StatusFound)
}
