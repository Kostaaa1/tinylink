package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/authcontext"
	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/domain/auth"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/Kostaaa1/tinylink/internal/middleware"
	"github.com/Kostaaa1/tinylink/pkg/errorhandler"
	"github.com/Kostaaa1/tinylink/pkg/validator"
	"github.com/gorilla/mux"
)

type TinylinkHandler struct {
	errorhandler.ErrorHandler
	service *tinylink.Service
}

func NewTinylinkHandler(tinylinkService *tinylink.Service, errHandler errorhandler.ErrorHandler) TinylinkHandler {
	return TinylinkHandler{
		ErrorHandler: errHandler,
		service:      tinylinkService,
	}
}

func (h TinylinkHandler) RegisterRoutes(r *mux.Router, auth middleware.Auth) {
	tinylinkRouter := r.PathPrefix("/tinylink").Subrouter()
	tinylinkRouter.Use(auth.Middleware)
	tinylinkRouter.HandleFunc("", h.Update).Methods("PATCH")
	tinylinkRouter.HandleFunc("", h.List).Methods("GET")
	tinylinkRouter.HandleFunc("/{alias}", h.Delete).Methods("DELETE")

	protectedRouter := r.PathPrefix("").Subrouter()
	protectedRouter.Use(auth.Middleware)
	protectedRouter.HandleFunc("/p/{alias:[a-zA-Z0-9]+}", h.Redirect).Methods("GET")

	r.HandleFunc("/{alias:[a-zA-Z0-9]+}", h.Redirect).Methods("GET")
	r.HandleFunc("/tinylink/create", h.Create).Methods("POST")
}

func (h TinylinkHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	claims := authcontext.GetClaims(ctx)

	links, err := h.service.List(ctx, claims)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, links, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h TinylinkHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req tinylink.UpdateTinylinkRequest
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

	claims := authcontext.GetClaims(ctx)

	tl, err := h.service.Update(ctx, claims, req)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNotFound):
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

func (h TinylinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req tinylink.InsertTinylinkRequest
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

	claims, err := auth.GetClaimsFromRequest(r)
	if err != nil {
		h.UnauthorizedResponse(w, r)
		return
	}

	tl, err := h.service.Insert(ctx, claims, req)
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

func (h TinylinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	alias := mux.Vars(r)["alias"]

	var tl *tinylink.Tinylink
	var err error

	if strings.HasPrefix(r.URL.Path, "/p/") {
		claims := authcontext.ClaimsFromCtx(ctx)
		if claims == nil {
			h.UnauthorizedResponse(w, r)
			return
		}

		tl, err = h.service.GetPersonal(ctx, claims, alias)
		if err != nil {
			h.ServerErrorResponse(w, r, err)
			return
		}
	} else {
		tl, err = h.service.Get(ctx, alias)
		if err != nil {
			h.ServerErrorResponse(w, r, err)
			return
		}
	}

	w.Header().Set("Location", tl.OriginalURL)
	w.WriteHeader(http.StatusFound)
}

func (h TinylinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	claims := authcontext.ClaimsFromCtx(ctx)

	alias := mux.Vars(r)["alias"]
	if err := h.service.Delete(r.Context(), claims, alias); err != nil {
		switch {
		case errors.Is(err, data.ErrNotFound):
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
