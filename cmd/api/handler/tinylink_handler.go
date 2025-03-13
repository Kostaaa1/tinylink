package handler

import (
	"net/http"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/middleware/session"
	"github.com/Kostaaa1/tinylink/internal/services"
	"github.com/Kostaaa1/tinylink/internal/validator"
	"github.com/gorilla/mux"
)

type TinylinkHandler struct {
	*ErrorHandler
	service *services.TinylinkService
}

func NewTinylinkHandler(tinylinkService *services.TinylinkService, errHandler *ErrorHandler) *TinylinkHandler {
	return &TinylinkHandler{
		ErrorHandler: errHandler,
		service:      tinylinkService,
	}
}

func (h *TinylinkHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/getAll", h.List).Methods("GET")
	r.HandleFunc("/create", h.Save).Methods("POST")
	r.HandleFunc("/{alias}", h.Redirect).Methods("GET")
	r.HandleFunc("/{alias}", h.Delete).Methods("DELETE")
}

func (h *TinylinkHandler) List(w http.ResponseWriter, r *http.Request) {
	sessionID, err := session.GetID(r)
	if err != nil {
		h.BadRequestResponse(w, r, err)
		return
	}

	links, err := h.service.List(r.Context(), sessionID)
	if err != nil {
		h.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"data": links}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h *TinylinkHandler) Save(w http.ResponseWriter, r *http.Request) {
	var req data.CreateTinylinkRequest

	if err := readJSON(r, &req); err != nil {
		h.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.New()
	if ok := req.IsValid(v); !ok {
		h.FailedValidationResponse(w, r, v.Errors)
		return
	}

	sessionID, err := session.GetID(r)
	if err != nil {
		h.BadRequestResponse(w, r, err)
		return
	}

	tl, err := h.service.Save(r.Context(), sessionID, req.URL, req.Alias)
	if err != nil {
		status, msg := h.MapErrorToStatus(err)
		h.ErrorResponse(w, r, status, msg)
		return
	}

	if err := writeJSON(w, http.StatusCreated, envelope{"data": tl}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h *TinylinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	sessionID, err := session.GetID(r)
	if err != nil {
		h.BadRequestResponse(w, r, err)
		return
	}

	tinylinkAlias := mux.Vars(r)["alias"]
	tl, err := h.service.Get(r.Context(), sessionID, tinylinkAlias)
	if err != nil {
		return
	}

	w.Header().Set("Location", tl.OriginalURL)
	w.WriteHeader(http.StatusFound)
}

func (h *TinylinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	sessionID, err := session.GetID(r)
	if err != nil {
		h.BadRequestResponse(w, r, err)
		return
	}

	tinylink := mux.Vars(r)["alias"]

	if err := h.service.Delete(r.Context(), sessionID, tinylink); err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"msg": "tinylink succesfully deleted"}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}
