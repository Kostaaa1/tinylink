package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Kostaaa1/tinylink/internal/application/interfaces"
	"github.com/Kostaaa1/tinylink/internal/errors"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/middleware/session"
	"github.com/Kostaaa1/tinylink/internal/interface/utils/jsonutil"
	"github.com/gorilla/mux"
)

type TinylinkHandler struct {
	service interfaces.TinylinkService
}

func NewTinylinkHandler(r *mux.Router, tinylinkService interfaces.TinylinkService) {
	h := TinylinkHandler{
		service: tinylinkService,
	}
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("tettttst\n"))
	}).Methods("GET")
	r.HandleFunc("/getAll", h.List).Methods("GET")
	r.HandleFunc("/create", h.Create).Methods("POST")
	r.HandleFunc("/{alias}", h.Redirect).Methods("GET")
	r.HandleFunc("/{alias}", h.Delete).Methods("DELETE")
}

func (h *TinylinkHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	sessionID, err := session.GetID(r)
	if err != nil {
		errors.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	links, err := h.service.List(ctx, sessionID)
	if err != nil {
		errors.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	fmt.Println("write links: ", links)
	if err := jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"data": links}, nil); err != nil {
		errors.ServerErrorResponse(w, r, err)
	}
}

func (h *TinylinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL   string `json:"url"`
		Alias string `json:"alias"`
	}
	if err := jsonutil.ReadJSON(r, &input); err != nil {
		errors.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// validate input
	_, err := url.Parse(input.URL)
	if err != nil {
		errors.ErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("invalid url format for: %s", input.URL))
		return
	}

	sessionID, err := session.GetID(r)
	if err != nil {
		errors.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	ctx := context.Background()

	tl, err := h.service.Save(ctx, sessionID, input.URL, input.Alias)
	if err != nil {
		fmt.Println("failed to create tinylink.", err)
		errors.ServerErrorResponse(w, r, err)
		return
	}

	if err := jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"data": tl}, nil); err != nil {
		errors.ServerErrorResponse(w, r, err)
	}
}

func (h *TinylinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	sessionID, err := session.GetID(r)
	if err != nil {
		errors.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	tinylinkAlias := mux.Vars(r)["alias"]
	ctx := context.Background()

	tl, err := h.service.Get(ctx, sessionID, tinylinkAlias)
	if err != nil {
		return
	}
	w.Header().Set("Location", tl.OriginalURL)
	w.WriteHeader(http.StatusFound)
}

func (h *TinylinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	sessionID, err := session.GetID(r)
	if err != nil {
		errors.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	tinylink := mux.Vars(r)["alias"]

	ctx := context.Background()

	if err := h.service.Delete(ctx, sessionID, tinylink); err != nil {
		return
	}

	if err := jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"msg": "tinylink succesfully deleted"}, nil); err != nil {
		// a.serverErrorResponse(w, r, err)
	}
}
