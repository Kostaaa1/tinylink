package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

func (h *TinylinkHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/getAll", h.List).Methods("GET")
	r.HandleFunc("/create", h.Create).Methods("POST")
	r.HandleFunc("/{alias}", h.Redirect).Methods("GET")
	r.HandleFunc("/{alias}", h.Delete).Methods("DELETE")
}

func (h *TinylinkHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	user := auth.UserFromCtx(ctx)
	// userID := user.GetID()
	// if userID == "" {
	// 	h.UnauthorizedResponse(w, r)
	// 	return
	// }
	fmt.Println(user)
	userID := user.GetID()

	links, err := h.service.List(ctx, userID)
	if err != nil {
		h.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"data": links}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h *TinylinkHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	if !auth.IsAuthenticated(ctx) {
		h.UnauthorizedResponse(w, r)
		return
	}

	tl, err := h.service.Create(ctx, req.URL, req.Alias)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrURLExists):
			h.ErrorResponse(w, r, http.StatusConflict, "Tinylink already exists for this URL")
		case errors.Is(err, data.ErrAliasExists):
			h.ErrorResponse(w, r, http.StatusConflict, "Alias not available")
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusCreated, envelope{"data": tl}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h *TinylinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	user := auth.UserFromCtx(ctx)
	userId := user.GetID()
	if userId == "" {
		h.UnauthorizedResponse(w, r)
		return
	}

	alias := mux.Vars(r)["alias"]
	tl, err := h.service.Get(r.Context(), userId, alias)
	if err != nil {
		return
	}

	w.Header().Set("Location", tl.URL)
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
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"msg": "tinylink succesfully deleted"}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}
