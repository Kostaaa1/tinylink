package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Kostaaa1/tinylink/internal/application/interfaces"
	errResp "github.com/Kostaaa1/tinylink/internal/errors"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/middleware/session"
	"github.com/Kostaaa1/tinylink/internal/interface/dto/request"
	"github.com/Kostaaa1/tinylink/internal/interface/utils/jsonutil"
	"github.com/Kostaaa1/tinylink/internal/validator"
	"github.com/gorilla/mux"
)

type TinylinkHandler struct {
	service interfaces.TinylinkService
}

func NewTinylinkHandler(r *mux.Router, tinylinkService interfaces.TinylinkService) {
	h := TinylinkHandler{
		service: tinylinkService,
	}
	r.HandleFunc("/getAll", h.List).Methods("GET")
	r.HandleFunc("/create", h.Save).Methods("POST")
	r.HandleFunc("/{alias}", h.Redirect).Methods("GET")
	r.HandleFunc("/{alias}", h.Delete).Methods("DELETE")
}

func (h *TinylinkHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	sessionID, err := session.GetID(r)
	if err != nil {
		errResp.BadRequestResponse(w, r, err)
		return
	}

	links, err := h.service.List(ctx, sessionID)
	if err != nil {
		errResp.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"data": links}, nil); err != nil {
		errResp.ServerErrorResponse(w, r, err)
	}
}

func (h *TinylinkHandler) Save(w http.ResponseWriter, r *http.Request) {
	var req request.CreateTinylinkRequest

	if err := jsonutil.ReadJSON(r, &req); err != nil {
		errResp.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.New()
	if ok := req.IsValid(v); !ok {
		errResp.FailedValidationResponse(w, r, v.Errors)
		return
	}

	sessionID, err := session.GetID(r)
	if err != nil {
		errResp.BadRequestResponse(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	tl, err := h.service.Save(ctx, sessionID, req.URL, req.Alias)
	if err != nil {
		status, msg := errResp.MapErrorToStatus(err)
		errResp.ErrorResponse(w, r, status, msg)
		return
	}

	if err := jsonutil.WriteJSON(w, http.StatusCreated, jsonutil.Envelope{"data": tl}, nil); err != nil {
		errResp.ServerErrorResponse(w, r, err)
	}
}

func (h *TinylinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	sessionID, err := session.GetID(r)
	if err != nil {
		errResp.BadRequestResponse(w, r, err)
		return
	}

	tinylinkAlias := mux.Vars(r)["alias"]

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

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
		errResp.BadRequestResponse(w, r, err)
		return
	}

	tinylink := mux.Vars(r)["alias"]

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.service.Delete(ctx, sessionID, tinylink); err != nil {
		errResp.ServerErrorResponse(w, r, err)
		return
	}

	writeJSONResponse(w, r, http.StatusOK, jsonutil.Envelope{"msg": "tinylink succesfully deleted"}, nil)
}

func writeJSONResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}, headers http.Header) {
	if err := jsonutil.WriteJSON(w, status, jsonutil.Envelope{"data": data}, headers); err != nil {
		errResp.ServerErrorResponse(w, r, err)
	}
}
