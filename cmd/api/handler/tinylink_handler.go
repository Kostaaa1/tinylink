package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/middleware/auth"
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

func (h *TinylinkHandler) RegisterRoutes(r *mux.Router, authMW func(http.Handler) http.Handler) {
	protected := r.PathPrefix("").Subrouter()
	protected.Use(authMW)
	protected.HandleFunc("/tinylink", h.Update).Methods("PUT")
	protected.HandleFunc("/{alias}", h.Delete).Methods("DELETE")
	protected.HandleFunc("/tinylink", h.Create).Methods("POST")
	protected.HandleFunc("/tinylink", h.List).Methods("GET")
	r.HandleFunc("/{alias}", h.Redirect).Methods("GET")
	// this should not be protected
}

func (h *TinylinkHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	user := auth.UserFromCtx(ctx)
	userID := user.GetID()

	links, err := h.service.List(ctx, userID)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"data": links}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h *TinylinkHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req data.UpdateTinylinkRequest
	if err := readJSON(r, &req); err != nil {
		h.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if ok := req.IsValid(v); !ok {
		h.FailedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	tl, err := h.service.Update(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.NotFoundResponse(w, r)
		case errors.Is(err, data.ErrURLExists):
			h.ErrorResponse(w, r, http.StatusConflict, err.Error())
		case errors.Is(err, data.ErrAliasExists):
			h.ErrorResponse(w, r, http.StatusConflict, err.Error())
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"data": tl}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h *TinylinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req data.InsertTinylinkRequest
	if err := readJSON(r, &req); err != nil {
		h.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if ok := req.IsValid(v); !ok {
		h.FailedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	tl, err := h.service.Insert(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrURLExists):
			h.ErrorResponse(w, r, http.StatusConflict, err.Error())
		case errors.Is(err, data.ErrAliasExists):
			h.ErrorResponse(w, r, http.StatusConflict, err.Error())
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	w.Header().Set("Location", strconv.FormatUint(tl.ID, 10))
	if err := writeJSON(w, http.StatusCreated, envelope{"data": tl}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h *TinylinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	alias := mux.Vars(r)["alias"]
	tl, err := h.service.Get(ctx, alias)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	w.Header().Set("Location", tl.OriginalURL)
	w.WriteHeader(http.StatusFound)
}

func (h *TinylinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	user := auth.UserFromCtx(ctx)
	userId := user.GetID()
	if userId == "" {
		h.UnauthorizedResponse(w, r)
		return
	}

	alias := mux.Vars(r)["alias"]
	if err := h.service.Delete(r.Context(), userId, alias); err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.NotFoundResponse(w, r)
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"msg": "tinylink succesfully deleted"}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}
