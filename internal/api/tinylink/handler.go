package tinylink

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Kostaaa1/tinylink/internal/constants"
	"github.com/Kostaaa1/tinylink/internal/domain/auth"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/Kostaaa1/tinylink/pkg/errhandler"
	"github.com/Kostaaa1/tinylink/pkg/jsonutil"
	"github.com/Kostaaa1/tinylink/pkg/validator"
	"github.com/gorilla/mux"
)

type TinylinkHandler struct {
	errhandler.ErrorHandler
	service *tinylink.Service
	log     *slog.Logger
}

func NewTinylinkHandler(service *tinylink.Service, errHandler errhandler.ErrorHandler, log *slog.Logger) TinylinkHandler {
	return TinylinkHandler{
		ErrorHandler: errHandler,
		service:      service,
		log:          log,
	}
}

func (h TinylinkHandler) RegisterRoutes(r *mux.Router, protected mux.MiddlewareFunc) {
	// protectedRoute.HandleFunc("/p/{alias:[a-zA-Z0-9]+}", h.Redirect).Methods("GET")
	protectedTL := r.PathPrefix("/tinylink").Subrouter()
	protectedTL.Use(protected)
	protectedTL.HandleFunc("/bulk-insert", h.BulkInsert).Methods("POST")
	protectedTL.HandleFunc("/{alias}", h.Delete).Methods("DELETE")
	protectedTL.HandleFunc("", h.Update).Methods("PATCH")
	protectedTL.HandleFunc("/list", h.List).Methods("GET")
	r.HandleFunc("/{alias:[a-zA-Z0-9]+}", h.Redirect).Methods("GET")
	r.HandleFunc("/tinylink/create", h.Create).Methods("POST")
}

func (h TinylinkHandler) BulkInsert(w http.ResponseWriter, r *http.Request) {
	h.log.Info("bulk insert")
	fmt.Fprint(w, "bulk insert")
}

// List all Tinylinks for user/guest
// @Summary List Tinylinks
// @Description Retrieves a list of Tinylinks for the authenticated user. This is protected route. Access tokey is needed
// @Tags Tinylink
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} tinylink.Tinylink
// @Failure 401 {object} jsonutil.Response
// @Failure 500 {object} jsonutil.Response
// @Router /tinylink/list [get]
func (h TinylinkHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	userCtx := auth.FromContext(ctx)

	links, err := h.service.List(ctx, userCtx)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := jsonutil.Write(w, http.StatusOK, links, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h TinylinkHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateTinylinkRequest
	if err := jsonutil.Read(r, &req); err != nil {
		h.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if _ = req.Validate(v); !v.Valid() {
		h.FailedValidationResponse(w, r, v.Errors)
		return
	}

	userCtx := auth.FromContext(r.Context())
	if !userCtx.IsAuthenticated {
		h.UnauthorizedResponse(w, r)
		return
	}

	params := tinylink.UpdateTinylinkParams{
		ID:      req.ID,
		URL:     req.URL,
		Alias:   req.Alias,
		Domain:  req.Domain,
		Private: req.Private,
		UserID:  *userCtx.UserID,
	}

	tl, err := h.service.Update(r.Context(), params)
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrNotFound):
			h.NotFoundResponse(w, r)
		case errors.Is(err, tinylink.ErrAliasExists):
			h.ErrorResponse(w, r, http.StatusConflict, err.Error())
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	if err := jsonutil.Write(w, http.StatusOK, tl, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h TinylinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTinylinkRequest
	err := jsonutil.Read(r, &req)
	if err != nil {
		h.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if _ = req.Validate(v); !v.Valid() {
		h.FailedValidationResponse(w, r, v.Errors)
		return
	}

	sig := auth.GetUserSignature(r)
	if sig.UserID == nil && sig.GuestUUID == nil {
		h.UnauthorizedResponse(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	params := tinylink.CreateTinylinkParams{
		URL:       req.URL,
		Alias:     req.Alias,
		Domain:    req.Domain,
		Private:   req.Private,
		UserID:    sig.UserID,
		GuestUUID: *sig.GuestUUID,
	}

	tl, err := h.service.Create(ctx, params)
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrUnauthenticated):
			h.UnauthorizedResponse(w, r)
		case errors.Is(err, tinylink.ErrAliasExists):
			h.ErrorResponse(w, r, http.StatusConflict, err.Error())
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	w.Header().Set("Location", strconv.FormatUint(tl.ID, 10))
	if err := jsonutil.Write(w, http.StatusCreated, tl, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h TinylinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	alias := mux.Vars(r)["alias"]

	var err error
	var url string

	// isPrivateRoute := strings.HasPrefix(r.URL.Path, "/p/")
	claims, _ := auth.ClaimsFromRequest(r)

	_, url, err = h.service.Redirect(r.Context(), &claims.UserID, alias)
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrNotFound):
			h.NotFoundResponse(w, r)
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusFound)
}

func (h TinylinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	alias := mux.Vars(r)["alias"]
	if alias == "" {
		h.BadRequestResponse(w, r, errors.New("missing alias in parameters"))
		return
	}

	userCtx := auth.FromContext(r.Context())
	if !userCtx.IsAuthenticated {
		h.UnauthorizedResponse(w, r)
		return
	}

	if err := h.service.Delete(r.Context(), *userCtx.UserID, alias); err != nil {
		switch {
		case errors.Is(err, constants.ErrNotFound):
			h.NotFoundResponse(w, r)
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	if err := jsonutil.Write(w, http.StatusOK, "tinylink succesfully deleted", nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}
