package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/auth"
	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/common/validator"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/gorilla/mux"
)

type TinylinkHandler struct {
	*ErrorHandler
	service *tinylink.Service
}

func NewTinylinkHandler(tinylinkService *tinylink.Service, errHandler *ErrorHandler) *TinylinkHandler {
	return &TinylinkHandler{
		ErrorHandler: errHandler,
		service:      tinylinkService,
	}
}

func (h *TinylinkHandler) RegisterRoutes(r *mux.Router) {
	tinylinkRouter := r.PathPrefix("/tinylink").Subrouter()
	tinylinkRouter.Use(auth.Middleware)
	tinylinkRouter.HandleFunc("", h.Update).Methods("PATCH")
	tinylinkRouter.HandleFunc("", h.Create).Methods("POST")
	tinylinkRouter.HandleFunc("", h.List).Methods("GET")
	tinylinkRouter.HandleFunc("/{alias}", h.Delete).Methods("DELETE")
	r.HandleFunc("/{alias:[a-zA-Z0-9]+}", h.Redirect).Methods("GET")
}

func (h *TinylinkHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	claims := auth.ClaimsFromCtx(ctx)

	links, err := h.service.List(ctx, claims.ID)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, links, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

type UpdateTinylinkRequest struct {
	ID      uint64 `json:"id"`
	Alias   string `json:"alias"`
	Private bool   `json:"private"`
	Domain  string `json:"domain"`
}

func (req *UpdateTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.ID != 0, "id", "must be provided")
	v.Check(req.Alias != "", "alias", "must be provided")
	v.Check(!(req.Alias != "" && len(req.Alias) < 5), "alias", "must be at least 5 characters long")
	return v.Valid()
}

func (h *TinylinkHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateTinylinkRequest
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

	tl, err := h.service.Update(ctx, req.ID, req.Alias, req.Domain, req.Private)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.NotFoundResponse(w, r)
		case errors.Is(err, tinylink.ErrURLExists):
			h.ErrorResponse(w, r, http.StatusConflict, err.Error())
		case errors.Is(err, tinylink.ErrAliasExists):
			h.ErrorResponse(w, r, http.StatusConflict, err.Error())
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, tl, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

type InsertTinylinkRequest struct {
	OriginalURL string `json:"original_url"`
	Alias       string `json:"alias"`
	Domain      string `json:"domain"`
	Private     bool   `json:"private"`
}

func (req *InsertTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.OriginalURL != "", "url", "must be provided")
	_, err := url.ParseRequestURI(req.OriginalURL)
	v.Check(err == nil, "url", "invalid url format")
	v.Check(!(req.Alias != "" && len(req.Alias) < 5), "alias", "must be at least 5 characters long")
	return v.Valid()
}

func (h *TinylinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req InsertTinylinkRequest
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

	tl, err := h.service.Insert(ctx, req.Alias, req.OriginalURL, req.Domain, req.Private)
	if err != nil {
		switch {
		case errors.Is(err, tinylink.ErrURLExists):
			h.ErrorResponse(w, r, http.StatusConflict, err.Error())
		case errors.Is(err, tinylink.ErrAliasExists):
			h.ErrorResponse(w, r, http.StatusConflict, err.Error())
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	w.Header().Set("Location", strconv.FormatUint(tl.ID, 10))
	if err := writeJSON(w, http.StatusCreated, tl, nil); err != nil {
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
	// ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	// defer cancel()

	// user := auth.UserFromCtx(ctx)
	// userId := user.GetID()
	// if userId == "" {
	// 	h.UnauthorizedResponse(w, r)
	// 	return
	// }

	// alias := mux.Vars(r)["alias"]
	// if err := h.service.Delete(r.Context(), userId, alias); err != nil {
	// 	switch {
	// 	case errors.Is(err, data.ErrRecordNotFound):
	// 		h.NotFoundResponse(w, r)
	// 	default:
	// 		h.ServerErrorResponse(w, r, err)
	// 	}
	// 	return
	// }

	// if err := writeJSON(w, http.StatusOK, envelope{"msg": "tinylink succesfully deleted"}, nil); err != nil {
	// 	h.ServerErrorResponse(w, r, err)
	// }
}
