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
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
	"github.com/Kostaaa1/tinylink/internal/middleware"
	"github.com/Kostaaa1/tinylink/pkg/errorhandler"
	"github.com/Kostaaa1/tinylink/pkg/validator"
	"github.com/gorilla/mux"
)

type TinylinkHandler struct {
	errorhandler.ErrorHandler
	service *tinylink.Service
}

type TinylinkDTO struct {
	ID          uint64     `json:"-"`
	Alias       string     `json:"alias"`
	OriginalURL string     `json:"original_url"`
	UserID      uint64     `json:"user_id,omitempty"`
	Private     bool       `json:"private"`
	UsageCount  int        `json:"usage_count"`
	Domain      string     `json:"domain,omitempty"`
	Version     uint64     `json:"version"`
	CreatedAt   time.Time  `json:"created_at"`
	LastVisited *time.Time `json:"last_visited"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

func toDTO(tl *tinylink.Tinylink) TinylinkDTO {
	var userID uint64
	if tl.UserID != nil {
		userID, _ = strconv.ParseUint(*tl.UserID, 10, 64)
	}

	var domain string
	if tl.Domain != nil {
		domain = *tl.Domain
	}

	var lastVisited *time.Time
	if tl.LastVisited > 0 {
		t := time.Unix(tl.LastVisited, 0)
		lastVisited = &t
	}

	var expiresAt *time.Time
	if tl.ExpiresAt > 0 {
		t := time.Unix(tl.ExpiresAt, 0)
		expiresAt = &t
	}

	return TinylinkDTO{
		ID:          tl.ID,
		Alias:       tl.Alias,
		OriginalURL: tl.OriginalURL,
		UserID:      userID,
		Private:     tl.Private,
		UsageCount:  tl.UsageCount,
		Domain:      domain,
		Version:     tl.Version,
		CreatedAt:   time.Unix(tl.CreatedAt, 0),
		LastVisited: lastVisited,
		ExpiresAt:   expiresAt,
	}
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
	tinylinkRouter.HandleFunc("/list", h.List).Methods("GET")
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
	claims := authcontext.Claims(ctx)

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
	claims := authcontext.Claims(ctx)

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
	claims, _ := token.ClaimsFromRequest(r)

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
	if err := writeJSON(w, http.StatusCreated, toDTO(tl), nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h TinylinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	alias := mux.Vars(r)["alias"]

	var err error
	var originalURL string

	if strings.HasPrefix(r.URL.Path, "/p/") {
		claims := authcontext.ClaimsFromCtx(ctx)
		if claims == nil {
			h.UnauthorizedResponse(w, r)
			return
		}
		originalURL, err = h.service.RedirectPersonal(ctx, claims, alias)
		if err != nil {
			h.ServerErrorResponse(w, r, err)
			return
		}
	} else {
		originalURL, err = h.service.Redirect(ctx, alias)
		if err != nil {
			h.ServerErrorResponse(w, r, err)
			return
		}
	}

	w.Header().Set("Location", originalURL)
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
