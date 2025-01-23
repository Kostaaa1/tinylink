// package main

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/url"

// 	"github.com/gorilla/mux"
// )

// type TinylinkHandler struct {
// 	service TinylinkService
// }

// func NewTinylinkHandler(r *mux.Router, tinylinkService TinylinkService) {
// 	h := TinylinkHandler{
// 		service: tinylinkService,
// 	}
// 	r.HandleFunc("/getAll", h.List).Methods("GET")
// 	r.HandleFunc("/create", h.Create).Methods("POST")
// 	r.HandleFunc("/{alias}", h.Redirect).Methods("GET")
// 	r.HandleFunc("/{alias}", h.Delete).Methods("DELETE")
// }

// func (h *TinylinkHandler) List(w http.ResponseWriter, r *http.Request) {
// 	ctx := context.Background()
// 	sessionID, err := getSessionID(r)
// 	if err != nil {
// 		return
// 	}

// 	links, err := h.service.List(ctx, sessionID)
// 	if err != nil {
// 		return
// 	}

// 	fmt.Println("write links: ", links)
// }

// func (h *TinylinkHandler) Create(w http.ResponseWriter, r *http.Request) {
// 	var input struct {
// 		URL   string `json:"url"`
// 		Alias string `json:"alias"`
// 	}
// 	if err := a.readJSON(r, &input); err != nil {
// 		// a.serverErrorResponse(w, r, err)
// 		return
// 	}

// 	// validate input
// 	_, err := url.Parse(input.URL)
// 	if err != nil {
// 		return
// 	}

// 	sessionID, err := getSessionID(r)
// 	if err != nil {
// 		return
// 	}

// 	ctx := context.Background()
// 	tl, err := h.service.Create(ctx, sessionID, input.URL, input.Alias)
// 	if err != nil {
// 		return
// 	}

// 	fmt.Println("write tlJ", tl)
// }

// func (h *TinylinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
// 	sessionID, err := getSessionID(r)
// 	if err != nil {
// 		return
// 	}

// 	tinylinkAlias := mux.Vars(r)["alias"]
// 	ctx := context.Background()

// 	tl, err := h.service.Get(ctx, sessionID, tinylinkAlias)
// 	if err != nil {
// 		return
// 	}
// 	w.Header().Set("Location", tl.OriginalURL)
// 	w.WriteHeader(http.StatusFound)
// }

// func (h *TinylinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
// 	sessionID, err := getSessionID(r)
// 	if err != nil {
// 		return
// 	}

// 	tinylink := mux.Vars(r)["alias"]

// 	ctx := context.Background()

// 	if err := h.service.Delete(ctx, sessionID, tinylink); err != nil {
// 		return
// 	}

// 	// if err := a.writeJSON(w, http.StatusOK, envelope{"msg": "tinylink succesfully deleted"}, nil); err != nil {
// 	// 	a.serverErrorResponse(w, r, err)
// 	// }
// }
